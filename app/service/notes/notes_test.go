package notes

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/service/eval"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const tmpl = `From: {{.From}}, To: {{.To}}, Date: {{.Date.Format "2006-01-02T15:04:05Z07:00"}}, Extras: {{.Extras}}
{{range .Categories}}{{.Title}}
{{ range .PRs }}- {{.Title}} ([#{{.Number}}]({{ .URL }}), branch {{ .SourceBranch }}) by @{{.Author.Username}} at {{ .ClosedAt }}
{{end}}
{{end}}`

const example = `From: 123, To: 456, Date: 2020-01-01T00:00:00Z, Extras: map[foo:bar]
Features
- Add feature 1 ([#3](url3), branch feat/feature-1) by @user3 at 2020-01-01 00:00:00 +0000 UTC
- Add feature 2 ([#2](url2), branch feat/feature-2) by @user2 at 2020-01-01 00:00:00 +0000 UTC

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
		SortField:      "-number",
		Template:       tmpl,
		UnusedTitle:    "Unused",
		IgnoreLabels:   []string{"ignore"},
		IgnoreBranchRe: regexp.MustCompile(`^ignore/`),
	}, &eval.Evaluator{}, map[string]string{"foo": "bar"})
	require.NoError(t, err)

	svc.now = func() time.Time { return tm }

	req := BuildRequest{
		From: "123", To: "456",
		ClosedPRs: []git.PullRequest{
			{
				Number:       2,
				Title:        "Add feature 2",
				ClosedAt:     tm,
				Author:       git.User{Username: "user2"},
				URL:          "url2",
				SourceBranch: "feat/feature-2",
				Labels:       []string{"feature"},
			},
			{
				Number:       1,
				Title:        "Fix bug 1",
				ClosedAt:     tm,
				Author:       git.User{Username: "user1"},
				URL:          "url1",
				SourceBranch: "fix/bug-1",
				Labels:       []string{"bug"},
			},
			{
				Number:       3,
				Title:        "Add feature 1",
				ClosedAt:     tm,
				Author:       git.User{Username: "user3"},
				URL:          "url3",
				SourceBranch: "feat/feature-1",
				Labels:       []string{"feature"},
			},
			{
				Number:       5,
				Title:        "Add feature 3",
				ClosedAt:     tm,
				Author:       git.User{Username: "user5"},
				URL:          "url5",
				SourceBranch: "blah/feature-3",
				Labels:       nil,
			},
			{
				Number:         7,
				Title:          "Ignore me",
				Body:           "ignore",
				Author:         git.User{Username: "user7"},
				Labels:         []string{"ignore"},
				ClosedAt:       tm,
				SourceBranch:   "feature/blah",
				URL:            "url7",
				ReceivedBySHAs: []string{"123"},
			},
		},
	}

	txt, err := svc.Build(context.Background(), req)
	require.NoError(t, err)

	assert.Equal(t, example, txt)
}

func TestBuilder_sortPRs(t *testing.T) {
	tm := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name  string
		field string
		prs   []git.PullRequest
		want  []git.PullRequest
	}{
		{
			name:  "sort by number",
			field: "number",
			prs:   []git.PullRequest{{Number: 2}, {Number: 1}},
			want:  []git.PullRequest{{Number: 1}, {Number: 2}},
		},
		{
			name:  "sort by number desc",
			field: "-number",
			prs:   []git.PullRequest{{Number: 1}, {Number: 2}},
			want:  []git.PullRequest{{Number: 2}, {Number: 1}},
		},
		{
			name:  "sort by title",
			field: "title",
			prs:   []git.PullRequest{{Title: "b"}, {Title: "a"}},
			want:  []git.PullRequest{{Title: "a"}, {Title: "b"}},
		},
		{
			name:  "sort by title desc",
			field: "-title",
			prs:   []git.PullRequest{{Title: "a"}, {Title: "b"}},
			want:  []git.PullRequest{{Title: "b"}, {Title: "a"}},
		},
		{
			name:  "sort by closed at",
			field: "closed",
			prs:   []git.PullRequest{{ClosedAt: tm.Add(time.Hour)}, {ClosedAt: tm}},
			want:  []git.PullRequest{{ClosedAt: tm}, {ClosedAt: tm.Add(time.Hour)}},
		},
		{
			name:  "sort by closed at desc",
			field: "-closed",
			prs:   []git.PullRequest{{ClosedAt: tm}, {ClosedAt: tm.Add(time.Hour)}},
			want:  []git.PullRequest{{ClosedAt: tm.Add(time.Hour)}, {ClosedAt: tm}},
		},
		{
			name:  "sort by author",
			field: "author",
			prs: []git.PullRequest{
				{Author: git.User{Username: "author2"}, Number: 2},
				{Author: git.User{Username: "author2"}, Number: 1},
				{Author: git.User{Username: "author1"}},
			},
			want: []git.PullRequest{
				{Author: git.User{Username: "author1"}},
				{Author: git.User{Username: "author2"}, Number: 1},
				{Author: git.User{Username: "author2"}, Number: 2},
			},
		},
		{
			name:  "sort by author desc",
			field: "-author",
			prs: []git.PullRequest{
				{Author: git.User{Username: "author1"}},
				{Author: git.User{Username: "author2"}, Number: 2},
				{Author: git.User{Username: "author2"}, Number: 1},
			},
			want: []git.PullRequest{
				{Author: git.User{Username: "author2"}, Number: 1},
				{Author: git.User{Username: "author2"}, Number: 2},
				{Author: git.User{Username: "author1"}},
			},
		},
		{
			name:  "default",
			field: "",
			prs:   []git.PullRequest{{Number: 2}, {Number: 1}},
			want:  []git.PullRequest{{Number: 1}, {Number: 2}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			(&Builder{Config: Config{SortField: tt.field}}).sortPRs(tt.prs)
			assert.Equal(t, tt.want, tt.prs)
		})
	}
}
