package eval

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/git/engine"
	gengine "github.com/Semior001/releaseit/app/git/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"text/template"
)

func TestGit_previousTag(t *testing.T) {
	t.Run("last tag", func(t *testing.T) {
		eng := &gengine.InterfaceMock{
			ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
				return []git.Tag{{Name: "v0.2.0"}, {Name: "v0.1.0"}}, nil
			},
		}

		res := execGitTmpl(t, eng, `{{ previousTag "v0.2.0" .Tags }}`,
			struct{ Tags []string }{Tags: []string{"v0.2.0", "v0.1.0"}})
		assert.Equal(t, "v0.1.0", res)
	})

	t.Run("first tag", func(t *testing.T) {
		eng := &gengine.InterfaceMock{
			ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
				return []git.Tag{{Name: "v0.1.0"}}, nil
			},
		}

		res := execGitTmpl(t, eng, `{{ previousTag "v0.1.0" .Tags }}`,
			struct{ Tags []string }{Tags: []string{"v0.2.0", "v0.1.0"}},
		)
		assert.Equal(t, "HEAD", res)
	})

	t.Run("tag's commit SHA", func(t *testing.T) {
		svc := &gengine.InterfaceMock{
			ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
				return []git.Tag{
					{Name: "v0.2.0", Commit: git.Commit{SHA: "sha"}},
					{Name: "v0.1.0"},
				}, nil
			},
		}

		res := execGitTmpl(t, svc, `{{ previousTag "sha" .Tags }}`,
			struct{ Tags []string }{Tags: []string{"v0.2.0", "v0.1.0"}},
		)
		assert.Equal(t, "v0.1.0", res)
	})

	t.Run("arbitrary commit SHA", func(t *testing.T) {
		eng := &gengine.InterfaceMock{
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
		}

		res := execGitTmpl(t, eng, `{{ previousTag "sha" .Tags }}`,
			struct{ Tags []string }{Tags: []string{"v0.2.0", "v0.1.0"}},
		)
		assert.Equal(t, "v0.1.0", res)
	})

	t.Run("nothing found", func(t *testing.T) {
		eng := &gengine.InterfaceMock{
			ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
				return []git.Tag{{Name: "v0.1.0"}}, nil
			},
			CompareFunc: func(ctx context.Context, from, to string) (git.CommitsComparison, error) {
				assert.Equal(t, "sha", to)
				assert.Equal(t, "v0.1.0", from)
				return git.CommitsComparison{}, nil
			},
		}

		res := execGitTmpl(t, eng, `{{ previousTag "sha" .Tags }}`,
			struct{ Tags []string }{Tags: []string{"v0.2.0", "v0.1.0"}},
		)
		assert.Equal(t, "HEAD", res)
	})
}

func TestGit_lastCommit(t *testing.T) {
	eng := &gengine.InterfaceMock{
		GetLastCommitOfBranchFunc: func(ctx context.Context, branch string) (string, error) {
			assert.Equal(t, "master", branch)
			return "sha", nil
		},
	}

	res := execGitTmpl(t, eng, `{{ lastCommit "master" }}`, nil)
	assert.Equal(t, "sha", res)
}

func TestGit_tags(t *testing.T) {
	eng := &gengine.InterfaceMock{
		ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
			return []git.Tag{{Name: "v0.1.0"}, {Name: "v0.2.0"}}, nil
		},
	}

	res := execGitTmpl(t, eng, `{{ tags }}`, nil)
	assert.Equal(t, fmt.Sprintf("%v", []string{"v0.2.0", "v0.1.0"}), res)
}

func TestGit_headed(t *testing.T) {
	res := execGitTmpl(t, nil, `{{ headed .List }}`, struct{ List []string }{List: []string{"v0.1.0"}})
	assert.Equal(t, fmt.Sprintf("%v", []string{"HEAD", "v0.1.0"}), res)
}

func TestGit_prTitles(t *testing.T) {
	res := execGitTmpl(t, nil, `{{ prTitles .List }}`, struct{ List []git.PullRequest }{
		List: []git.PullRequest{
			{Title: "title1"},
			{Title: "title2"},
			{Title: "title3"},
		},
	})
	assert.Equal(t, fmt.Sprintf("%v", []string{"title1", "title2", "title3"}), res)
}

func execGitTmpl(t *testing.T, eng engine.Interface, expr string, data any) string {
	fns, err := (&Git{Engine: eng}).Funcs(context.Background())
	require.NoError(t, err)

	tmpl, err := template.New("").Funcs(fns).Parse(expr)
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	require.NoError(t, tmpl.Execute(buf, data))

	return buf.String()
}

func TestGit_String(t *testing.T) {
	assert.Equal(t, "git", (&Git{}).String())
}
