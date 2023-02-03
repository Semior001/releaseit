// Package notes wraps engine interfaces with common logic
// unrelated to any particular engine implementation.
package notes

import (
	"fmt"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/Semior001/releaseit/app/git"
	"github.com/samber/lo"
)

const defaultTemplate = `Version {{.To}}
{{if not .Categories}}- No changes{{end}}{{range .Categories}}{{.Title}}
{{range .PRs}}- {{.Title}} (#{{.Number}}) by @{{.Author}}{{end}}
{{end}}`

// Builder provides methods to form changelog.
type Builder struct {
	Config
	Extras map[string]string

	now  func() time.Time
	tmpl *template.Template
}

// NewBuilder creates a new Builder.
func NewBuilder(cfg Config, extras map[string]string) (*Builder, error) {
	svc := &Builder{Extras: extras, Config: cfg, now: time.Now}

	if svc.Template == "" {
		svc.Template = defaultTemplate
	}

	tmpl, err := template.New("changelog").
		Funcs(lo.OmitByKeys(sprig.FuncMap(), []string{"env", "expandenv"})).
		Parse(svc.Template)
	if err != nil {
		return nil, fmt.Errorf("parsing template: %w", err)
	}

	svc.tmpl = tmpl

	return svc, nil
}

// BuildRequest is a request for changelog building.
type BuildRequest struct {
	From      string
	To        string
	ClosedPRs []git.PullRequest
}

// Build builds the changelog for the tag.
func (s *Builder) Build(req BuildRequest) (string, error) {
	data := tmplData{
		From:   req.From,
		To:     req.To,
		Date:   s.now(),
		Extras: s.Extras,
	}

	usedPRs := make([]bool, len(req.ClosedPRs))

	for _, category := range s.Categories {
		categoryData := categoryTmplData{Title: category.Title}

		for i, pr := range req.ClosedPRs {
			if len(lo.Intersect(pr.Labels, s.IgnoreLabels)) > 0 {
				usedPRs[i] = true
				continue
			}

			hasBranchPrefix := category.BranchRe != nil && category.BranchRe.MatchString(pr.Branch)
			hasAnyOfLabels := len(lo.Intersect(pr.Labels, category.Labels)) > 0

			if hasAnyOfLabels || hasBranchPrefix {
				usedPRs[i] = true
				categoryData.PRs = append(categoryData.PRs, prTmplData{
					Number:         pr.Number,
					Title:          pr.Title,
					Author:         pr.Author.Username,
					ClosedAt:       pr.ClosedAt,
					URL:            pr.URL,
					Branch:         pr.Branch,
					ReceivedBySHAs: pr.ReceivedBySHAs,
				})
			}
		}

		if len(categoryData.PRs) == 0 {
			continue
		}

		s.sortPRs(categoryData.PRs)
		data.Categories = append(data.Categories, categoryData)
	}

	if s.UnusedTitle != "" {
		if unlabeled := s.makeUnlabeledCategory(usedPRs, req.ClosedPRs); len(unlabeled.PRs) > 0 {
			s.sortPRs(unlabeled.PRs)
			data.Categories = append(data.Categories, unlabeled)
		}
	}

	buf := &strings.Builder{}

	if err := s.tmpl.Execute(buf, data); err != nil {
		return "", fmt.Errorf("executing template for changelog: %w", err)
	}

	return buf.String(), nil
}

func (s *Builder) makeUnlabeledCategory(used []bool, prs []git.PullRequest) categoryTmplData {
	category := categoryTmplData{Title: s.UnusedTitle}

	for i, pr := range prs {
		if used[i] {
			continue
		}

		category.PRs = append(category.PRs, prTmplData{
			Number:   pr.Number,
			Title:    pr.Title,
			Author:   pr.Author.Username,
			URL:      pr.URL,
			ClosedAt: pr.ClosedAt,
			Branch:   pr.Branch,
		})
	}

	return category
}

func (s *Builder) sortPRs(prs []prTmplData) {
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
				return prs[i].Author > prs[j].Author
			}
			if prs[i].Author == prs[j].Author {
				return prs[i].Number < prs[j].Number
			}
			return prs[i].Author < prs[j].Author
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
	From       string
	To         string
	Categories []categoryTmplData
	Date       time.Time // always set to the time when the changelog is generated
	Extras     map[string]string
}

type categoryTmplData struct {
	Title string
	PRs   []prTmplData
}

type prTmplData struct {
	Number         int
	Title          string
	Author         string
	URL            string
	Branch         string
	ClosedAt       time.Time
	ReceivedBySHAs []string
}
