// Package service provides the core functionality of the application.
package service

import (
	"context"
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

	req := notes.BuildRequest{FromSHA: from, ToSHA: to, ClosedPRs: prs}

	text, err := s.ReleaseNotesBuilder.Build(req)
	if err != nil {
		return fmt.Errorf("build release notes: %w", err)
	}

	if err = s.Notifier.Send(ctx, to, text); err != nil {
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
		"previous_tag": func(alias string) (string, error) {
			tags, err := s.Engine.ListTags(ctx)
			if err != nil {
				return "", fmt.Errorf("list tags: %w", err)
			}

			// if by any chance alias is a tag itself
			for idx, tag := range tags {
				if tag.Name == alias || tag.Commit.SHA == alias {
					if idx+1 == len(tags) {
						return "HEAD", nil
					}

					return tags[idx+1].Name, nil
				}
			}

			// otherwise, we find the closest tag
			for _, tag := range tags {
				comp, err := s.Engine.Compare(ctx, tag.Name, alias)
				if err != nil {
					return "", fmt.Errorf("compare tag %s with commit %s: %w", tag.Commit.SHA, alias, err)
				}

				if len(comp.Commits) > 0 {
					return tag.Name, nil
				}
			}

			return "HEAD", nil
		},
	}
}

func (s *Service) evalCommitIDs(ctx context.Context, fromExpr, toExpr string) (from string, to string, err error) {
	data := struct{ From, To string }{From: fromExpr, To: toExpr}

	evalID := func(expr string) (string, error) {
		tmpl, err := template.New("").Funcs(s.exprFuncs(ctx)).Parse(expr)
		if err != nil {
			return "", fmt.Errorf("parse template: %w", err)
		}

		res := &strings.Builder{}
		if err = tmpl.Execute(res, data); err != nil {
			return "", fmt.Errorf("execute template: %w", err)
		}

		return res.String(), nil
	}

	if from, err = evalID(fromExpr); err != nil {
		return "", "", fmt.Errorf("evaluate 'from' expression: %w", err)
	}

	if to, err = evalID(toExpr); err != nil {
		return "", "", fmt.Errorf("evaluate 'to' expression: %w", err)
	}

	return from, to, nil
}
