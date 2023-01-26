// Package service provides the core functionality of the application.
package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/git/engine"
	"github.com/Semior001/releaseit/app/notify"
	"github.com/Semior001/releaseit/app/service/notes"
	"github.com/samber/lo"
)

// Service wraps repository storage and services
type Service struct {
	Engine                engine.Interface
	ReleaseNotesBuilder   *notes.Builder
	Notifier              notify.Destination
	SquashCommitMessageRx *regexp.Regexp
}

// ReleaseBetween makes a release between two commit SHAs.
func (s *Service) ReleaseBetween(ctx context.Context, from, to string) error {
	from, err := s.getCommitSHA(ctx, from)
	if err != nil {
		return fmt.Errorf("get 'from' commit SHA: %w", err)
	}

	if to, err = s.getCommitSHA(ctx, to); err != nil {
		return fmt.Errorf("get 'to' commit SHA: %w", err)
	}

	prs, err := s.closedPRsBetweenSHA(ctx, from, to)
	if err != nil {
		return fmt.Errorf("get closed pull requests between %s and %s: %w", from, to, err)
	}

	req := notes.BuildRequest{
		Version:   fmt.Sprintf("%s..%s", from, to),
		FromSHA:   from,
		ToSHA:     to,
		ClosedPRs: prs,
	}

	text, err := s.ReleaseNotesBuilder.Build(req)
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
		from = "HEAD"
	} else {
		// otherwise use the previous tag (tags sorted in descending order of creation)
		from = tags[tagIdx+1].Commit.SHA
	}

	prs, err := s.closedPRsBetweenSHA(ctx, from, to)
	if err != nil {
		return fmt.Errorf("get closed pull requests between %s and %s: %w", from, to, err)
	}

	req := notes.BuildRequest{
		Version:   fmt.Sprintf("%s..%s", from, to),
		FromSHA:   from,
		ToSHA:     to,
		ClosedPRs: prs,
	}

	text, err := s.ReleaseNotesBuilder.Build(req)
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
		refCommitSHA, ok := s.isMergeCommit(commit)
		if !ok {
			continue
		}

		prs, err := s.Engine.ListPRsOfCommit(ctx, refCommitSHA)
		if err != nil {
			return nil, fmt.Errorf("list pull requests of commit %s: %w", refCommitSHA, err)
		}

		for _, pr := range prs {
			if !pr.ClosedAt.IsZero() {
				res = append(res, pr)
			}
		}
	}

	return lo.UniqBy(res, func(item git.PullRequest) int { return item.Number }), nil
}

func (s *Service) isMergeCommit(commit git.Commit) (prAttachedSHA string, ok bool) {
	if len(commit.ParentSHAs) > 1 {
		return commit.ParentSHAs[1], true
	}

	if s.SquashCommitMessageRx.MatchString(commit.Message) {
		return commit.SHA, true
	}

	return "", false
}

func (s *Service) exprFuncs(ctx context.Context) template.FuncMap {
	return template.FuncMap{
		"last_commit": func(branch string) (string, error) {
			return s.Engine.GetLastCommitOfBranch(ctx, branch)
		},
		"head": func() string { return "HEAD" },
	}
}

func (s *Service) getCommitSHA(ctx context.Context, expr string) (string, error) {
	const exprPrefix = "!!"
	if !strings.HasPrefix(expr, exprPrefix) {
		return expr, nil
	}

	tmpl, err := template.New("").Funcs(s.exprFuncs(ctx)).Parse(strings.TrimPrefix(expr, exprPrefix))
	if err != nil {
		return "", fmt.Errorf("parse expression: %w", err)
	}

	res := &strings.Builder{}

	if err = tmpl.Execute(res, nil); err != nil {
		return "", fmt.Errorf("execute expression: %w", err)
	}

	return res.String(), nil
}
