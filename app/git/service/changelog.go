// Package service wraps engine interfaces with common logic
// unrelated to any particular engine implementation.
package service

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/Semior001/releaseit/app/git"
)

const defaultTemplate = `Version {{.Tag}}
{{if not .Categories}}- No changes{{end}}{{range .Categories}}{{.Title}}
{{range .PRs}}- {{.Title}} (#{{.Number}}) by @{{.Author}}{{end}}
{{end}}`

// ReleaseNotesBuilder provides methods to form changelog.
type ReleaseNotesBuilder struct {
	changelogTmpl *template.Template
	categories    []Category
	ignoreLabels  []string
	unusedTitle   string
	sortField     string
}

// Category describes pull request category with its title,
// which will be derived to template and labels, that indicates
// the belonging to this category.
type Category struct {
	Title  string
	Labels []string
}

// Params specifies parameters needed
// to initialize ReleaseNotesBuilder.
type Params struct {
	Template     string
	IgnoreLabels []string
	Categories   []Category
	UnusedTitle  string
	SortField    string
}

// NewChangelogBuilder makes new service from the specified parameters.
func NewChangelogBuilder(params Params) (*ReleaseNotesBuilder, error) {
	if params.Template == "" {
		params.Template = defaultTemplate
	}

	tmpl, err := template.New("changelog").Parse(params.Template)
	if err != nil {
		return nil, fmt.Errorf("parse changelogTmpl: %w", err)
	}

	return &ReleaseNotesBuilder{
		changelogTmpl: tmpl,
		categories:    params.Categories,
		ignoreLabels:  params.IgnoreLabels,
		sortField:     params.SortField,
		unusedTitle:   params.UnusedTitle,
	}, nil
}

type changelogTmplData struct {
	Tag        string
	Categories []categoryTmplData
	Date       time.Time
}

type categoryTmplData struct {
	Title string
	PRs   []prTmplData
}

type prTmplData struct {
	Number int
	Title  string
	Author string
	Closed time.Time
}

// Build builds the changelog for the tag.
func (s *ReleaseNotesBuilder) Build(cl git.Changelog) (string, error) {
	// building template data
	data := changelogTmplData{Tag: cl.Tag.Name, Date: time.Now()}

	usedPRs := make([]bool, len(cl.ClosedPRs))

	for _, category := range s.categories {
		categoryData := categoryTmplData{Title: category.Title}

		for i, pr := range cl.ClosedPRs {
			if containsOneOf(pr.Labels, s.ignoreLabels) {
				usedPRs[i] = true
				continue
			}

			if containsOneOf(pr.Labels, category.Labels) {
				usedPRs[i] = true
				categoryData.PRs = append(categoryData.PRs, prTmplData{
					Number: pr.Number,
					Title:  pr.Title,
					Author: pr.Author.Username,
					Closed: pr.ClosedAt,
				})
			}
		}

		if len(categoryData.PRs) == 0 {
			continue
		}

		s.sortPRs(categoryData.PRs)
		data.Categories = append(data.Categories, categoryData)
	}

	if s.unusedTitle != "" {
		if unlabeled := s.makeUnlabeledCategory(usedPRs, cl.ClosedPRs); len(unlabeled.PRs) > 0 {
			s.sortPRs(unlabeled.PRs)
			data.Categories = append(data.Categories, unlabeled)
		}
	}

	buf := &bytes.Buffer{}

	if err := s.changelogTmpl.Execute(buf, data); err != nil {
		return "", fmt.Errorf("executing template for changelog: %w", err)
	}

	return buf.String(), nil
}

func (s *ReleaseNotesBuilder) makeUnlabeledCategory(used []bool, prs []git.PullRequest) categoryTmplData {
	category := categoryTmplData{Title: s.unusedTitle}

	for i, pr := range prs {
		if used[i] {
			continue
		}

		category.PRs = append(category.PRs, prTmplData{
			Number: pr.Number,
			Title:  pr.Title,
			Author: pr.Author.Username,
		})
	}

	return category
}

func (s *ReleaseNotesBuilder) sortPRs(prs []prTmplData) {
	sort.Slice(prs, func(i, j int) bool {
		switch s.sortField {
		case "+number", "-number", "number":
			if strings.HasPrefix(s.sortField, "-") {
				return prs[i].Number > prs[j].Number
			}
			return prs[i].Number < prs[j].Number
		case "+author", "-author", "author":
			if strings.HasPrefix(s.sortField, "-") {
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
			if strings.HasPrefix(s.sortField, "-") {
				return prs[i].Title > prs[j].Title
			}
			return prs[i].Title < prs[j].Title
		case "+closed", "-closed", "closed":
			if strings.HasPrefix(s.sortField, "-") {
				return prs[i].Closed.After(prs[j].Closed)
			}
			return prs[i].Closed.Before(prs[j].Closed)
		default:
			return prs[i].Number < prs[j].Number
		}
	})
}

func containsOneOf(arr []string, entries []string) bool {
	for _, m := range arr {
		for _, entry := range entries {
			if m == entry {
				return true
			}
		}
	}
	return false
}
