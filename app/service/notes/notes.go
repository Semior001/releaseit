// Package notes wraps engine interfaces with common logic
// unrelated to any particular engine implementation.
package notes

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/service/eval"
	"github.com/samber/lo"
)

// Builder provides methods to form changelog.
type Builder struct {
	Config
	Evaluator *eval.Evaluator
	Extras    map[string]string

	now func() time.Time
}

// NewBuilder creates a new Builder.
func NewBuilder(cfg Config, eval *eval.Evaluator, extras map[string]string) (*Builder, error) {
	svc := &Builder{
		Extras:    extras,
		Evaluator: eval,
		Config:    cfg,
		now:       time.Now,
	}

	return svc, nil
}

// BuildRequest is a request for changelog building.
type BuildRequest struct {
	From      string
	To        string
	ClosedPRs []git.PullRequest
	Commits   []git.Commit
}

// Build builds the changelog for the tag.
func (s *Builder) Build(ctx context.Context, req BuildRequest) (string, error) {
	data := tmplData{
		From:         req.From,
		To:           req.To,
		Date:         s.now(),
		Extras:       s.Extras,
		Total:        len(req.ClosedPRs),
		TotalCommits: len(req.Commits),
	}

	usedPRs := make([]bool, len(req.ClosedPRs))
	usedCommits := make([]bool, len(req.Commits))
	commitIDxBySHA := make(map[string]int, len(req.Commits))
	for i, commit := range req.Commits {
		commitIDxBySHA[commit.SHA] = i
	}

	for _, category := range s.Categories {
		categoryData := categoryTmplData{Title: category.Title}

		for i, pr := range req.ClosedPRs {
			if len(lo.Intersect(pr.Labels, s.IgnoreLabels)) > 0 {
				usedPRs[i] = true
				continue
			}

			hasBranchPrefix := category.BranchRe != nil && category.BranchRe.MatchString(pr.SourceBranch)
			hasAnyOfLabels := len(lo.Intersect(pr.Labels, category.Labels)) > 0

			if !hasAnyOfLabels && !hasBranchPrefix {
				continue
			}

			usedPRs[i] = true
			categoryData.PRs = append(categoryData.PRs, pr)
			for _, commit := range pr.ReceivedBySHAs {
				if commitIdx, ok := commitIDxBySHA[commit]; ok {
					usedCommits[commitIdx] = true
				}
			}
		}

		s.sortPRs(categoryData.PRs)
		data.Categories = append(data.Categories, categoryData)
	}

	for categoryIdx, category := range s.Categories {
		if category.CommitMsgRe == nil {
			continue
		}

		for idx, commit := range req.Commits {
			if usedCommits[idx] {
				continue
			}

			if !category.CommitMsgRe.MatchString(commit.Message) {
				continue
			}

			usedCommits[idx] = true
			data.Categories[categoryIdx].Commits = append(data.Categories[categoryIdx].Commits, commit)
		}
	}

	if s.UnusedTitle != "" {
		unlabeled := s.makeUnlabeledCategory(usedPRs, usedCommits, req.ClosedPRs, req.Commits)
		if len(unlabeled.PRs) > 0 || len(unlabeled.Commits) > 0 {
			s.sortPRs(unlabeled.PRs)
			data.Categories = append(data.Categories, unlabeled)
		}
	}

	res, err := s.Evaluator.Evaluate(ctx, s.Template, data)
	if err != nil {
		return "", fmt.Errorf("executing template for changelog: %w", err)
	}

	return res, nil
}

func (s *Builder) makeUnlabeledCategory(
	usedPR, usedCommit []bool,
	prs []git.PullRequest,
	commits []git.Commit,
) categoryTmplData {
	category := categoryTmplData{Title: s.UnusedTitle}

	for i, pr := range prs {
		if usedPR[i] {
			continue
		}

		category.PRs = append(category.PRs, pr)
	}

	for i, commit := range commits {
		if usedCommit[i] {
			continue
		}

		category.Commits = append(category.Commits, commit)
	}

	return category
}

func (s *Builder) sortPRs(prs []git.PullRequest) {
	sort.Slice(prs, func(i, j int) bool {
		switch s.SortField {
		case "+number", "-number", "number":
			if strings.HasPrefix(s.SortField, "-") {
				return prs[i].Number > prs[j].Number
			}
			return prs[i].Number < prs[j].Number
		case "+author", "-author", "author":
			if strings.HasPrefix(s.SortField, "-") {
				if prs[i].Author == prs[j].Author {
					return prs[i].Number < prs[j].Number
				}
				return prs[i].Author.Username > prs[j].Author.Username
			}
			if prs[i].Author == prs[j].Author {
				return prs[i].Number < prs[j].Number
			}
			return prs[i].Author.Username < prs[j].Author.Username
		case "+title", "-title", "title":
			if strings.HasPrefix(s.SortField, "-") {
				return prs[i].Title > prs[j].Title
			}
			return prs[i].Title < prs[j].Title
		case "+closed", "-closed", "closed":
			if strings.HasPrefix(s.SortField, "-") {
				return prs[i].ClosedAt.After(prs[j].ClosedAt)
			}
			return prs[i].ClosedAt.Before(prs[j].ClosedAt)
		default:
			return prs[i].Number < prs[j].Number
		}
	})
}

type tmplData struct {
	From         string
	To           string
	Date         time.Time // always set to the time when the changelog is generated
	Extras       map[string]string
	Total        int // total number of PRs
	TotalCommits int // total number of commits
	Categories   []categoryTmplData
}

type categoryTmplData struct {
	Title   string
	PRs     []git.PullRequest
	Commits []git.Commit
}
