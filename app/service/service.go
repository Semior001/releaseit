package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/git/engine"
	"github.com/Semior001/releaseit/app/notify"
	"github.com/samber/lo"
)

// Service wraps repository storage and services
type Service struct {
	Engine              engine.Interface
	ReleaseNotesBuilder *ReleaseNotesBuilder
	Notifier            notify.Destination
}

// ReleaseBetween makes a release between two commit SHAs.
func (s *Service) ReleaseBetween(ctx context.Context, from, to string) error {
	prs, err := s.closedPRsBetweenSHA(ctx, from, to)
	if err != nil {
		return fmt.Errorf("get closed pull requests between %s and %s: %w", from, to, err)
	}

	text, err := s.ReleaseNotesBuilder.Build(fmt.Sprintf("%s...%s", from, to), prs)
	if err != nil {
		return fmt.Errorf("build release notes: %w", err)
	}

	if err = s.Notifier.Send(ctx, "", text); err != nil {
		return fmt.Errorf("notify: %w", err)
	}

	return nil
}

// ReleaseTag aggregates the changelog from the changes of the latest tag and its predecessor
// and notifies consumers via provided notifier.
func (s *Service) ReleaseTag(ctx context.Context, tagName string) error {
	// resolving last two tags of the repo
	tags, err := s.Engine.ListTags(ctx)
	if err != nil {
		return fmt.Errorf("list tags: %w", err)
	}

	_, tagIdx, ok := lo.FindIndexOf(tags, func(tag git.Tag) bool {
		return tag.Name == tagName || tag.Commit.SHA == tagName
	})
	if !ok {
		return errors.New("tag not found")
	}

	var from string
	to := tags[tagIdx].Commit.SHA

	if len(tags) == 1 {
		// if the given tag is the only tag in the repository, then fetch
		// changelog since HEAD commit
		if from, err = s.Engine.HeadCommit(ctx); err != nil {
			return fmt.Errorf("get head commit: %w", err)
		}
	} else {
		// otherwise use the previous tag (tags sorted in descending order of creation)
		from = tags[tagIdx+1].Commit.SHA
	}

	prs, err := s.closedPRsBetweenSHA(ctx, from, to)
	if err != nil {
		return fmt.Errorf("get closed pull requests between %s and %s: %w", from, to, err)
	}

	text, err := s.ReleaseNotesBuilder.Build(tagName, prs)
	if err != nil {
		return fmt.Errorf("build release notes: %w", err)
	}

	if err = s.Notifier.Send(ctx, tagName, text); err != nil {
		return fmt.Errorf("notify: %w", err)
	}

	return nil
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

	return lo.UniqBy(res, func(item git.PullRequest) int { return item.Number }), nil
}
