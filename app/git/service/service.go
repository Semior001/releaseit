package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/git/engine"
	"github.com/samber/lo"
)

// Service wraps repository storage and services
type Service struct {
	Engine   engine.Interface
	Notifier notifier
}

type notifier interface {
	Send(ctx context.Context, changelog git.Changelog) error
}

// NewService makes new instance of Service.
func NewService(engine engine.Interface, notifier notifier) *Service {
	return &Service{
		Engine:   engine,
		Notifier: notifier,
	}
}

// Release aggregates the changelog from the changes of the latest tag and its predecessor
// and notifies consumers via provided notifier.
func (s *Service) Release(ctx context.Context, tag string) error {
	cl, err := s.Changelog(ctx, tag)
	if err != nil {
		return fmt.Errorf("aggregate changelog: %w", err)
	}

	if err = s.Notifier.Send(ctx, cl); err != nil {
		return fmt.Errorf("notify: %w", err)
	}

	return nil
}

// LastTag returns the last tag in the repository.
func (s *Service) LastTag(ctx context.Context) (git.Tag, error) {
	tags, err := s.Engine.ListTags(ctx)
	if err != nil {
		return git.Tag{}, fmt.Errorf("list tags: %w", err)
	}

	if len(tags) == 0 {
		return git.Tag{}, errors.New("repository has no tags")
	}

	return tags[0], nil
}

// Changelog returns changelog of the specified tag name, i.e. all changes
// between the specified tag (provided by name or SHA) and its predecessor
// (e.g. previous tag or HEAD of the repo).
func (s *Service) Changelog(ctx context.Context, tagName string) (git.Changelog, error) {
	// resolving last two tags of the repo
	tags, err := s.Engine.ListTags(ctx)
	if err != nil {
		return git.Changelog{}, fmt.Errorf("list tags: %w", err)
	}

	_, tagIdx, ok := lo.FindIndexOf(tags, func(tag git.Tag) bool {
		return tag.Name == tagName || tag.Commit.SHA == tagName
	})
	if !ok {
		return git.Changelog{}, errors.New("tag not found")
	}

	var from string
	to := tags[tagIdx].Commit.SHA

	if len(tags) == 1 {
		// if the given tag is the only tag in the repository, then fetch
		// changelog since HEAD commit
		if from, err = s.Engine.HeadCommit(ctx); err != nil {
			return git.Changelog{}, fmt.Errorf("get head commit: %w", err)
		}
	} else {
		// otherwise use the previous tag (tags sorted in descending order of creation)
		from = tags[tagIdx+1].Commit.SHA
	}

	prs, err := s.closedPRsBetweenSHA(ctx, from, to)
	if err != nil {
		return git.Changelog{}, fmt.Errorf("get closed pull requests between %s and %s: %w", from, to, err)
	}

	return git.Changelog{
		Tag:       tags[0],
		ClosedPRs: unique(prs),
	}, nil
}

func (s *Service) closedPRsBetweenSHA(ctx context.Context, fromSHA, toSHA string) ([]git.PullRequest, error) {
	var res []git.PullRequest

	commits, err := s.Engine.Compare(ctx, fromSHA, toSHA)
	if err != nil {
		return nil, fmt.Errorf("compare commits between %s and %s: %w", fromSHA, toSHA, err)
	}

	for _, commit := range commits.Commits {
		// if commit has more than one parent - probably it's a merge
		// commit
		// FIXME: needs better guessing, doesn't work with squash commits
		if len(commit.ParentSHAs) > 1 {
			prs, err := s.Engine.ListPRsOfCommit(ctx, commit.ParentSHAs[1])
			if err != nil {
				return nil, fmt.Errorf("list pull requests of commit %s: %w", commit.ParentSHAs[1], err)
			}

			for _, pr := range prs {
				if !pr.ClosedAt.IsZero() {
					res = append(res, pr)
				}
			}
		}
	}

	return res, nil
}

// unique unifies pull using their numbers as keys.
func unique(prs []git.PullRequest) []git.PullRequest {
	set := map[int]git.PullRequest{}
	for _, pr := range prs {
		if _, ok := set[pr.Number]; !ok {
			set[pr.Number] = pr
		}
	}
	res := make([]git.PullRequest, 0, len(set))
	for _, pr := range set {
		res = append(res, pr)
	}
	return res
}
