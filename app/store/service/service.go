package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Semior001/releaseit/app/store"
	"github.com/Semior001/releaseit/app/store/engine"
)

// Service wraps repository storage and services
type Service struct {
	Engine   engine.Interface
	Notifier notifier
}

type notifier interface {
	Send(ctx context.Context, changelog store.Changelog) error
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
func (s *Service) Release(ctx context.Context) error {
	// fetch release data
	tag, err := s.LastTag(ctx)
	if err != nil {
		return fmt.Errorf("get last tag: %w", err)
	}

	cl, err := s.Changelog(ctx, tag.Name)
	if err != nil {
		return fmt.Errorf("aggregate changelog: %w", err)
	}

	if err = s.Notifier.Send(ctx, cl); err != nil {
		return fmt.Errorf("notify: %w", err)
	}

	return nil
}

// LastTag returns the last tag in the repository.
func (s *Service) LastTag(ctx context.Context) (store.Tag, error) {
	tags, err := s.Engine.ListTags(ctx)
	if err != nil {
		return store.Tag{}, fmt.Errorf("list tags: %w", err)
	}

	if len(tags) == 0 {
		return store.Tag{}, errors.New("repository has no tags")
	}

	return tags[0], nil
}

// Changelog returns changelog of the specified tag name, i.e. all changes
// between the specified tag and its predecessor (e.g. previous tag or HEAD
// of the repo)
func (s *Service) Changelog(ctx context.Context, tagName string) (store.Changelog, error) {
	// resolving last two tags of the repo
	tags, err := s.Engine.ListTags(ctx)
	if err != nil {
		return store.Changelog{}, fmt.Errorf("list tags: %w", err)
	}

	tagIdx := -1

	for i, tag := range tags {
		if tag.Name == tagName {
			tagIdx = i
		}
	}

	if tagIdx == -1 {
		return store.Changelog{}, errors.New("tag not found")
	}

	var from string
	to := tags[0].Commit.SHA

	if len(tags) == 1 {
		// if the given tag is the only tag in the repository, then fetch
		// changelog since HEAD commit
		if from, err = s.Engine.HeadCommit(ctx); err != nil {
			return store.Changelog{}, fmt.Errorf("get head commit: %w", err)
		}
	} else {
		// otherwise use the previous tag (tags sorted in descending order of creation)
		from = tags[tagIdx+1].Commit.SHA
	}

	prs, err := s.closedPRsBetweenSHA(ctx, from, to)
	if err != nil {
		return store.Changelog{}, fmt.Errorf("get closed pull requests between %s and %s: %w", from, to, err)
	}

	return store.Changelog{
		Tag:       tags[0],
		ClosedPRs: prs,
	}, nil
}

func (s *Service) closedPRsBetweenSHA(ctx context.Context, fromSHA, toSHA string) ([]store.PullRequest, error) {
	var res []store.PullRequest

	commits, err := s.Engine.Compare(ctx, fromSHA, toSHA)
	if err != nil {
		return nil, fmt.Errorf("compare commits between %s and %s: %w", fromSHA, toSHA, err)
	}

	for _, commit := range commits.Commits {
		// if commit has more than one parent - probably it's a merge
		// commit
		// FIXME: needs better guessing, doesn't work with squash commits
		if len(commit.Parents) > 1 {
			prs, err := s.Engine.ListPRsOfCommit(ctx, commit.Parents[1].SHA)
			if err != nil {
				return nil, fmt.Errorf("list pull requests of commit %s: %w", commit.Parents[1].SHA, err)
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
