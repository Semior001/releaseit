package notes

import (
	"regexp"
	"testing"
	"time"

	"github.com/Semior001/releaseit/app/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const tmpl = `FromSHA: {{.FromSHA}}, ToSHA: {{.ToSHA}}, Date: {{.Date.Format "2006-01-02T15:04:05Z07:00"}}, Extras: {{.Extras}}
{{range .Categories}}{{.Title}}
{{ range .PRs }}- {{.Title}} ([#{{.Number}}]({{ .URL }}), branch {{ .Branch }}) by @{{.Author}} at {{ .ClosedAt }}{{end}}
{{end}}`

const example = `FromSHA: 123, ToSHA: 456, Date: 2020-01-01T00:00:00Z, Extras: map[foo:bar]
Features
- Add feature 1 ([#3](url3), branch feat/feature-1) by @user3 at 2020-01-01 00:00:00 +0000 UTC
Bug fixes
- Fix bug 1 ([#1](url1), branch fix/bug-1) by @user1 at 2020-01-01 00:00:00 +0000 UTC
Unused
- Add feature 3 ([#5](url5), branch blah/feature-3) by @user5 at 2020-01-01 00:00:00 +0000 UTC
`

func TestBuilder_Build(t *testing.T) {
	tm := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	svc, err := NewBuilder(Config{
		Categories: []CategoryConfig{
			{
				Title:    "Features",
				Labels:   []string{"feature"},
				BranchRe: regexp.MustCompile(`^feature/`),
			},
			{Title: "Bug fixes", Labels: []string{"bug"}},
		},
		SortField:      "+number",
		Template:       tmpl,
		UnusedTitle:    "Unused",
		IgnoreLabels:   []string{"ignore"},
		IgnoreBranchRe: regexp.MustCompile(`^ignore/`),
	}, map[string]string{"foo": "bar"})
	require.NoError(t, err)

	svc.now = func() time.Time { return tm }

	req := BuildRequest{
		FromSHA: "123",
		ToSHA:   "456",
		ClosedPRs: []git.PullRequest{
			{
				Number:   1,
				Title:    "Fix bug 1",
				ClosedAt: tm,
				Author:   git.User{Username: "user1"},
				URL:      "url1",
				Branch:   "fix/bug-1",
				Labels:   []string{"bug"},
			},
			{
				Number:   3,
				Title:    "Add feature 1",
				ClosedAt: tm,
				Author:   git.User{Username: "user3"},
				URL:      "url3",
				Branch:   "feat/feature-1",
				Labels:   []string{"feature"},
			},
			{
				Number:   5,
				Title:    "Add feature 3",
				ClosedAt: tm,
				Author:   git.User{Username: "user5"},
				URL:      "url5",
				Branch:   "blah/feature-3",
				Labels:   nil,
			},
		},
	}

	txt, err := svc.Build(req)
	require.NoError(t, err)

	assert.Equal(t, example, txt)
}
