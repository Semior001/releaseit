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
	"github.com/Semior001/releaseit/app/service/notes"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Changelog(t *testing.T) {
	t.Run("expressions on commits", func(t *testing.T) {
		compareCalledErr := errors.New("compare called")
		svc := &Service{
			Engine: &engine.InterfaceMock{
				GetCommitFunc: func(ctx context.Context, sha string) (git.Commit, error) {
					return git.Commit{SHA: "sha"}, nil
				},
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
			},
		}

		err := svc.Changelog(context.Background(), `{{ last_commit "master" }}`, `{{ previous_tag "v0.2.0" }}`)
		assert.ErrorIs(t, err, compareCalledErr)
	})

	t.Run("commit SHAs provided", func(t *testing.T) {
		now := time.Now()
		buf := &strings.Builder{}

		svc := &Service{
			SquashCommitMessageRx: regexp.MustCompile(`^squash: (.*)$`),
			Engine: &engine.InterfaceMock{
				GetCommitFunc: func(ctx context.Context, sha string) (git.Commit, error) {
					return git.Commit{SHA: "from", ParentSHAs: []string{"parent", "pr1"}, Message: "Pull request #1"}, nil
				},
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
					case "pr1":
						return []git.PullRequest{
							{Number: 1, Title: "Pull request #1", Branch: "feature/1", ClosedAt: now},
							{Number: 2, Title: "Pull request #2", Labels: []string{"bug"}, ClosedAt: now},
							{Number: 3, Title: "Pull request #3"},
						}, nil
					case "pr2":
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
			},
			ReleaseNotesBuilder: lo.Must(notes.NewBuilder(notes.Config{
				Categories: []notes.CategoryConfig{
					{Title: "Features", BranchRe: regexp.MustCompile(`^feature/.*$`)},
					{Title: "Bug fixes", Labels: []string{"bug"}},
				},
				IgnoreLabels: []string{"ignore"},
				SortField:    "number",
				Template:     `{{range .Categories}}{{.Title}}:{{range .PRs}} {{.Number}}{{end}} {{end}}`,
				UnusedTitle:  "Unused",
			}, nil)),
			Notifier: &notify.WriterNotifier{Writer: buf, Name: "buf"},
		}

		err := svc.Changelog(context.Background(), "from", "to")
		require.NoError(t, err)

		require.Equal(t, "Features: 1 Bug fixes: 2 Unused: 4 ", buf.String())
	})
}

func TestService_exprFuncs(t *testing.T) {
	t.Run("test default values", func(t *testing.T) {
		svc := &Service{
			Engine: &engine.InterfaceMock{
				ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
					return []git.Tag{{Name: "v0.2.0"}, {Name: "v0.1.0"}}, nil
				},
			},
		}

		from, to, err := svc.evalCommitIDs(context.Background(), `{{ previous_tag .To }}`, `{{ last_tag }}`)
		require.NoError(t, err)
		assert.Equal(t, "v0.1.0", from)
		assert.Equal(t, "v0.2.0", to)
	})

	t.Run("last_tag", func(t *testing.T) {
		svc := &Service{
			Engine: &engine.InterfaceMock{
				ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
					return []git.Tag{{Name: "v0.2.0"}, {Name: "v0.1.0"}}, nil
				},
			},
		}

		fn := svc.exprFuncs(context.Background())["last_tag"]

		tags, err := fn.(func() (string, error))()
		require.NoError(t, err)
		assert.Equal(t, "v0.2.0", tags)
	})

	t.Run("tags", func(t *testing.T) {
		svc := &Service{Engine: &engine.InterfaceMock{
			ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
				return []git.Tag{{Name: "v0.1.0"}, {Name: "v0.2.0"}}, nil
			},
		}}

		fn := svc.exprFuncs(context.Background())["tags"]

		tags, err := fn.(func() ([]string, error))()
		require.NoError(t, err)
		assert.Equal(t, []string{"v0.1.0", "v0.2.0"}, tags)
	})

	t.Run("last_commit", func(t *testing.T) {
		svc := &Service{Engine: &engine.InterfaceMock{
			GetLastCommitOfBranchFunc: func(ctx context.Context, branch string) (string, error) {
				assert.Equal(t, "master", branch)
				return "sha", nil
			},
		}}

		fn := svc.exprFuncs(context.Background())["last_commit"]

		sha, err := fn.(func(string) (string, error))("master")
		require.NoError(t, err)
		assert.Equal(t, "sha", sha)
	})

	t.Run("previous_tag", func(t *testing.T) {
		t.Run("just a tag", func(t *testing.T) {
			svc := &Service{Engine: &engine.InterfaceMock{
				ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
					return []git.Tag{{Name: "v0.2.0"}, {Name: "v0.1.0"}}, nil
				},
			}}

			fn := svc.exprFuncs(context.Background())["previous_tag"]

			tag, err := fn.(func(string) (string, error))("v0.2.0")
			require.NoError(t, err)
			assert.Equal(t, "v0.1.0", tag)
		})

		t.Run("first tag", func(t *testing.T) {
			svc := &Service{Engine: &engine.InterfaceMock{
				ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
					return []git.Tag{{Name: "v0.1.0"}}, nil
				},
			}}

			fn := svc.exprFuncs(context.Background())["previous_tag"]
			tag, err := fn.(func(string) (string, error))("v0.1.0")
			require.NoError(t, err)
			assert.Equal(t, "HEAD", tag)
		})

		t.Run("tag's commit SHA", func(t *testing.T) {
			svc := &Service{Engine: &engine.InterfaceMock{
				ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
					return []git.Tag{
						{Name: "v0.2.0", Commit: git.Commit{SHA: "sha"}},
						{Name: "v0.1.0"},
					}, nil
				},
			}}

			fn := svc.exprFuncs(context.Background())["previous_tag"]
			tag, err := fn.(func(string) (string, error))("sha")
			require.NoError(t, err)
			assert.Equal(t, "v0.1.0", tag)
		})

		t.Run("arbitrary commit SHA", func(t *testing.T) {
			svc := &Service{Engine: &engine.InterfaceMock{
				ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
					return []git.Tag{
						{Name: "v0.2.0"},
						{Name: "v0.1.0"},
					}, nil
				},
				CompareFunc: func(ctx context.Context, from, to string) (git.CommitsComparison, error) {
					assert.Equal(t, "sha", to)
					if from == "v0.2.0" {
						return git.CommitsComparison{}, nil
					}

					assert.Equal(t, "v0.1.0", from)
					return git.CommitsComparison{Commits: []git.Commit{{SHA: "sha"}}}, nil
				},
			}}

			fn := svc.exprFuncs(context.Background())["previous_tag"]
			tag, err := fn.(func(string) (string, error))("sha")
			require.NoError(t, err)
			assert.Equal(t, "v0.1.0", tag)
		})

		t.Run("nothing found", func(t *testing.T) {
			svc := &Service{Engine: &engine.InterfaceMock{
				ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
					return []git.Tag{{Name: "v0.1.0"}}, nil
				},
				CompareFunc: func(ctx context.Context, from, to string) (git.CommitsComparison, error) {
					assert.Equal(t, "sha", to)
					assert.Equal(t, "v0.1.0", from)
					return git.CommitsComparison{}, nil
				},
			}}

			fn := svc.exprFuncs(context.Background())["previous_tag"]
			tag, err := fn.(func(string) (string, error))("sha")
			require.NoError(t, err)
			assert.Equal(t, "HEAD", tag)
		})
	})
}
