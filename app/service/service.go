// Package service provides the core functionality of the application.
package service

import (
	"context"
	"fmt"
	gengine "github.com/Semior001/releaseit/app/git/engine"
	"regexp"

	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/notify"
	"github.com/Semior001/releaseit/app/service/eval"
	"github.com/Semior001/releaseit/app/service/notes"
	"github.com/samber/lo"
)

// Service wraps repository storage and services
type Service struct {
	Engine                gengine.Interface
	Evaluator             *eval.Evaluator
	ReleaseNotesBuilder   *notes.Builder
	Notifier              notify.Destination
	SquashCommitMessageRx *regexp.Regexp
}

// Changelog makes a release between two commit SHAs.
func (s *Service) Changelog(ctx context.Context, fromExpr, toExpr string) error {
	from, to, err := s.evalCommitIDs(ctx, fromExpr, toExpr)
	if err != nil {
		return fmt.Errorf("evaluate commit IDs: %w", err)
	}

	prs, err := s.closedPRsBetweenSHA(ctx, from, to)
	if err != nil {
		return fmt.Errorf("get closed pull requests between %s and %s: %w", from, to, err)
	}

	req := notes.BuildRequest{From: from, To: to, ClosedPRs: prs}

	text, err := s.ReleaseNotesBuilder.Build(ctx, req)
	if err != nil {
		return fmt.Errorf("build release notes: %w", err)
	}

	if err = s.Notifier.Send(ctx, text); err != nil {
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
		if ok := s.isMergeCommit(commit); !ok {
			continue
		}

		prs, err := s.Engine.ListPRsOfCommit(ctx, commit.SHA)
		if err != nil {
			return nil, fmt.Errorf("list pull requests of commit %s: %w", commit.SHA, err)
		}

		for _, pr := range prs {
			if !pr.ClosedAt.IsZero() {
				pr.ReceivedBySHAs = append(pr.ReceivedBySHAs, commit.SHA)
				res = append(res, pr)
			}
		}
	}

	// merge "received by sha" between PRs
	uniqPRs := map[int]git.PullRequest{}
	for _, pr := range res {
		if prev, ok := uniqPRs[pr.Number]; ok {
			pr.ReceivedBySHAs = append(pr.ReceivedBySHAs, prev.ReceivedBySHAs...)
		}

		uniqPRs[pr.Number] = pr
	}

	return lo.Values(uniqPRs), nil
}

func (s *Service) isMergeCommit(commit git.Commit) bool {
	return len(commit.ParentSHAs) > 1 || s.SquashCommitMessageRx.MatchString(commit.Message)
}

func (s *Service) evalCommitIDs(ctx context.Context, fromExpr, toExpr string) (from string, to string, err error) {
	if to, err = s.Evaluator.Evaluate(ctx, toExpr, nil); err != nil {
		return "", "", fmt.Errorf("evaluate 'to' expression: %w", err)
	}

	if from, err = s.Evaluator.Evaluate(ctx, fromExpr, struct{ To string }{To: to}); err != nil {
		return "", "", fmt.Errorf("evaluate 'from' expression: %w", err)
	}

	if from == "" || to == "" {
		return "", "", fmt.Errorf("empty commit ID; from: %s, to: %s", from, to)
	}

	return from, to, nil
}
