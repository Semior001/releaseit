// Package service wraps engine interfaces with common logic
// unrelated to any particular engine implementation.
package service

import (
	"bytes"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/Semior001/releaseit/app/git"
	"github.com/samber/lo"
)

const defaultTemplate = `Version {{.Version}}
{{if not .Categories}}- No changes{{end}}{{range .Categories}}{{.Title}}
{{range .PRs}}- {{.Title}} (#{{.Number}}) by @{{.Author}}{{end}}
{{end}}`

// ReleaseNotesBuilder provides methods to form changelog.
type ReleaseNotesBuilder struct {
	Template     string
	Categories   []Category
	IgnoreLabels []string
	UnusedTitle  string
	SortField    string

	tmpl *template.Template
	once sync.Once
}

// Category describes pull request category with its title,
// which will be derived to template and labels, that indicates
// the belonging to this category.
type Category struct {
	Title        string
	Labels       []string
	BranchRegexp *regexp.Regexp
}

type changelogTmplData struct {
	Version    string
	Categories []categoryTmplData
	Date       time.Time
}

type categoryTmplData struct {
	Title string
	PRs   []prTmplData
}

type prTmplData struct {
	Number   int
	Title    string
	Author   string
	URL      string
	Branch   string
	ClosedAt time.Time
}

// Build builds the changelog for the tag.
func (s *ReleaseNotesBuilder) Build(version string, closedPRs []git.PullRequest) (string, error) {
	var err error
	s.once.Do(func() {
		if s.Template == "" {
			s.Template = defaultTemplate
		}
		s.tmpl, err = template.New("changelog").Funcs(funcs).Parse(s.Template)
	})
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	// building template data
	data := changelogTmplData{Version: version, Date: time.Now()}

	usedPRs := make([]bool, len(closedPRs))

	for _, category := range s.Categories {
		categoryData := categoryTmplData{Title: category.Title}

		for i, pr := range closedPRs {
			if len(lo.Intersect(pr.Labels, s.IgnoreLabels)) > 0 {
				usedPRs[i] = true
				continue
			}

			hasBranchPrefix := category.BranchRegexp != nil && category.BranchRegexp.MatchString(pr.Branch)
			hasAnyOfLabels := len(lo.Intersect(pr.Labels, category.Labels)) > 0

			if hasAnyOfLabels || hasBranchPrefix {
				usedPRs[i] = true
				categoryData.PRs = append(categoryData.PRs, prTmplData{
					Number:   pr.Number,
					Title:    pr.Title,
					Author:   pr.Author.Username,
					ClosedAt: pr.ClosedAt,
					URL:      pr.URL,
					Branch:   pr.Branch,
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
		if unlabeled := s.makeUnlabeledCategory(usedPRs, closedPRs); len(unlabeled.PRs) > 0 {
			s.sortPRs(unlabeled.PRs)
			data.Categories = append(data.Categories, unlabeled)
		}
	}

	buf := &bytes.Buffer{}

	if err := s.tmpl.Execute(buf, data); err != nil {
		return "", fmt.Errorf("executing template for changelog: %w", err)
	}

	return buf.String(), nil
}

func (s *ReleaseNotesBuilder) makeUnlabeledCategory(used []bool, prs []git.PullRequest) categoryTmplData {
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
		})
	}

	return category
}

func (s *ReleaseNotesBuilder) sortPRs(prs []prTmplData) {
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

var funcs = template.FuncMap{
	"time_LoadLocation": time.LoadLocation,
}
