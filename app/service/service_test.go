package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/git/engine"
	"github.com/Semior001/releaseit/app/notify"
	"github.com/Semior001/releaseit/app/service/eval"
	"github.com/Semior001/releaseit/app/service/notes"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Changelog(t *testing.T) {
	t.Run("expressions on commits", func(t *testing.T) {
		compareCalledErr := errors.New("compare called")

		eng := &engine.InterfaceMock{
			GetLastCommitOfBranchFunc: func(ctx context.Context, branch string) (string, error) {
				assert.Equal(t, "master", branch)
				return "sha", nil
			},
			ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
				return []git.Tag{{Name: "v0.2.0"}, {Name: "v0.1.0"}}, nil
			},
			CompareFunc: func(ctx context.Context, fromSHA string, toSHA string) (git.CommitsComparison, error) {
				assert.Equal(t, "sha", fromSHA)
				assert.Equal(t, "v0.1.0", toSHA)
				return git.CommitsComparison{}, compareCalledErr
			},
		}

		svc := &Service{
			Evaluator: &eval.Evaluator{Engine: eng},
			Engine:    eng,
		}

		err := svc.Changelog(context.Background(), `{{ last_commit "master" }}`, `{{ previous_tag "v0.2.0" }}`)
		assert.ErrorIs(t, err, compareCalledErr)
	})

	t.Run("commit SHAs provided", func(t *testing.T) {
		now := time.Now()
		buf := &strings.Builder{}

		eng := &engine.InterfaceMock{
			CompareFunc: func(ctx context.Context, from, to string) (git.CommitsComparison, error) {
				return git.CommitsComparison{
					Commits: []git.Commit{
						{SHA: "from", ParentSHAs: []string{"parent", "pr1"}, Message: "Pull request #1"},
						// this won't be picked up
						{SHA: "intermediate", ParentSHAs: []string{"from"}, Message: "intermediate commit message"},
						{SHA: "intermediate_squashed", ParentSHAs: []string{"intermediate"}, Message: "squash: Pull request #4"},
						{SHA: "to", ParentSHAs: []string{"intermediate", "pr2"}, Message: "Pull request #2"},
					},
					TotalCommits: 4,
				}, nil
			},
			ListPRsOfCommitFunc: func(ctx context.Context, sha string) ([]git.PullRequest, error) {
				switch sha {
				case "from":
					return []git.PullRequest{
						{Number: 1, Title: "Pull request #1", SourceBranch: "feature/1", ClosedAt: now},
						{Number: 2, Title: "Pull request #2", Labels: []string{"bug"}, ClosedAt: now},
						{Number: 3, Title: "Pull request #3"},
					}, nil
				case "to":
					return []git.PullRequest{
						{Number: 2, Title: "Pull request #2", Labels: []string{"bug"}, ClosedAt: now},
						{Number: 3, Title: "Pull request #5", Labels: []string{"ignore"}},
					}, nil
				case "intermediate_squashed":
					return []git.PullRequest{{Number: 4, Title: "Pull request #4", ClosedAt: now}}, nil
				default:
					return nil, fmt.Errorf("unhandled sha: %s", sha)
				}
			},
		}

		svc := &Service{
			SquashCommitMessageRx: regexp.MustCompile(`^squash: (.*)$`),
			Evaluator:             &eval.Evaluator{Engine: eng},
			Engine:                eng,
			ReleaseNotesBuilder: lo.Must(notes.NewBuilder(notes.Config{
				Categories: []notes.CategoryConfig{
					{Title: "Features", BranchRe: regexp.MustCompile(`^feature/.*$`)},
					{Title: "Bug fixes", Labels: []string{"bug"}},
				},
				IgnoreLabels: []string{"ignore"},
				SortField:    "number",
				Template:     `{{range .Categories}}{{.Title}}:{{range .PRs}} {{.Number}}{{end}} {{end}}`,
				UnusedTitle:  "Unused",
			}, &eval.Evaluator{}, nil)),
			Notifier: &notify.WriterNotifier{Writer: buf, Name: "buf"},
		}

		err := svc.Changelog(context.Background(), "from", "to")
		require.NoError(t, err)

		require.Equal(t, "Features: 1 Bug fixes: 2 Unused: 4 ", buf.String())
	})
}
