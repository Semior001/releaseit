// Package service provides the core functionality of the application.
package service

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sync"

	gengine "github.com/Semior001/releaseit/app/git/engine"
	"golang.org/x/sync/errgroup"

	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/notify"
	"github.com/Semior001/releaseit/app/service/eval"
	"github.com/Semior001/releaseit/app/service/notes"
	"github.com/samber/lo"
)

// Service wraps repository storage and services
type Service struct {
	Engine                  gengine.Interface
	Evaluator               *eval.Evaluator
	ReleaseNotesBuilder     *notes.Builder
	Notifier                notify.Destination
	FetchMergeCommitsFilter *regexp.Regexp
	MaxConcurrentPRRequests int
}

// Changelog makes a release between two commit SHAs.
func (s *Service) Changelog(ctx context.Context, fromExpr, toExpr string) error {
	log.Printf("[DEBUG] evaluating commit IDs from %s to %s", fromExpr, toExpr)
	from, to, err := s.evalCommitIDs(ctx, fromExpr, toExpr)
	if err != nil {
		return fmt.Errorf("evaluate commit IDs: %w", err)
	}

	log.Printf("[DEBUG] comparing commits between %s and %s", from, to)
	compare, err := s.Engine.Compare(ctx, from, to)
	if err != nil {
		return fmt.Errorf("compare commits between %s and %s: %w", from, to, err)
	}

	log.Printf("[DEBUG] got total of %d commits", len(compare.Commits))
	log.Printf("[DEBUG] aggregating closed pull requests between %s and %s", from, to)

	prs, err := s.closedPRsBetweenSHA(ctx, compare.Commits)
	if err != nil {
		return fmt.Errorf("get closed pull requests between %s and %s: %w", from, to, err)
	}

	req := notes.BuildRequest{From: from, To: to, ClosedPRs: prs, Commits: compare.Commits}

	log.Printf("[DEBUG] building release notes for %d pull requests", len(prs))
	text, err := s.ReleaseNotesBuilder.Build(ctx, req)
	if err != nil {
		return fmt.Errorf("build release notes: %w", err)
	}

	log.Printf("[DEBUG] sending release notes to destinations")
	if err = s.Notifier.Send(ctx, text); err != nil {
		return fmt.Errorf("notify: %w", err)
	}

	return nil
}

func (s *Service) closedPRsBetweenSHA(ctx context.Context, commits []git.Commit) ([]git.PullRequest, error) {
	var res []git.PullRequest

	ewg, ctx := errgroup.WithContext(ctx)

	if s.MaxConcurrentPRRequests < 1 {
		s.MaxConcurrentPRRequests = 1
	}

	ewg.SetLimit(s.MaxConcurrentPRRequests)
	mu := sync.Mutex{}

	for _, commit := range commits {
		commit := commit
		if ok := s.isMergeCommit(commit); !ok {
			continue
		}

		ewg.Go(func() error {
			prs, err := s.Engine.ListPRsOfCommit(ctx, commit.SHA)
			if err != nil {
				return fmt.Errorf("list pull requests of commit %s: %w", commit.SHA, err)
			}

			mu.Lock()
			defer mu.Unlock()

			for _, pr := range prs {
				if !pr.ClosedAt.IsZero() {
					pr.ReceivedBySHAs = append(pr.ReceivedBySHAs, commit.SHA)
					res = append(res, pr)
				}
			}

			return nil
		})
	}

	if err := ewg.Wait(); err != nil {
		return nil, fmt.Errorf("wait for pull requests: %w", err)
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
	return len(commit.ParentSHAs) > 1 || s.FetchMergeCommitsFilter.MatchString(commit.Message)
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
