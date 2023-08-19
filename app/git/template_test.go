package git

import (
	"bytes"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"text/template"
)

func TestTemplateFuncs_previousTag(t *testing.T) {
	t.Run("last tag", func(t *testing.T) {
		eng := &RepositoryMock{
			ListTagsFunc: func(ctx context.Context) ([]Tag, error) {
				return []Tag{{Name: "v0.2.0"}, {Name: "v0.1.0"}}, nil
			},
		}

		res := execTmpl(t, eng, `{{ previousTag "v0.2.0" .Tags }}`,
			struct{ Tags []string }{Tags: []string{"v0.2.0", "v0.1.0"}})
		assert.Equal(t, "v0.1.0", res)
	})

	t.Run("first tag", func(t *testing.T) {
		eng := &RepositoryMock{
			ListTagsFunc: func(ctx context.Context) ([]Tag, error) {
				return []Tag{{Name: "v0.1.0"}}, nil
			},
		}

		res := execTmpl(t, eng, `{{ previousTag "v0.1.0" .Tags }}`,
			struct{ Tags []string }{Tags: []string{"v0.2.0", "v0.1.0"}},
		)
		assert.Equal(t, "HEAD", res)
	})

	t.Run("tag's commit SHA", func(t *testing.T) {
		svc := &RepositoryMock{
			ListTagsFunc: func(ctx context.Context) ([]Tag, error) {
				return []Tag{
					{Name: "v0.2.0", Commit: Commit{SHA: "sha"}},
					{Name: "v0.1.0"},
				}, nil
			},
		}

		res := execTmpl(t, svc, `{{ previousTag "sha" .Tags }}`,
			struct{ Tags []string }{Tags: []string{"v0.2.0", "v0.1.0"}},
		)
		assert.Equal(t, "v0.1.0", res)
	})

	t.Run("arbitrary commit SHA", func(t *testing.T) {
		eng := &RepositoryMock{
			ListTagsFunc: func(ctx context.Context) ([]Tag, error) {
				return []Tag{
					{Name: "v0.2.0"},
					{Name: "v0.1.0"},
				}, nil
			},
			CompareFunc: func(ctx context.Context, from, to string) (CommitsComparison, error) {
				assert.Equal(t, "sha", to)
				if from == "v0.2.0" {
					return CommitsComparison{}, nil
				}

				assert.Equal(t, "v0.1.0", from)
				return CommitsComparison{Commits: []Commit{{SHA: "sha"}}}, nil
			},
		}

		res := execTmpl(t, eng, `{{ previousTag "sha" .Tags }}`,
			struct{ Tags []string }{Tags: []string{"v0.2.0", "v0.1.0"}},
		)
		assert.Equal(t, "v0.1.0", res)
	})

	t.Run("nothing found", func(t *testing.T) {
		eng := &RepositoryMock{
			ListTagsFunc: func(ctx context.Context) ([]Tag, error) {
				return []Tag{{Name: "v0.1.0"}}, nil
			},
			CompareFunc: func(ctx context.Context, from, to string) (CommitsComparison, error) {
				assert.Equal(t, "sha", to)
				assert.Equal(t, "v0.1.0", from)
				return CommitsComparison{}, nil
			},
		}

		res := execTmpl(t, eng, `{{ previousTag "sha" .Tags }}`,
			struct{ Tags []string }{Tags: []string{"v0.2.0", "v0.1.0"}},
		)
		assert.Equal(t, "HEAD", res)
	})
}

func TestTemplateFuncs_lastCommit(t *testing.T) {
	eng := &RepositoryMock{
		GetLastCommitOfBranchFunc: func(ctx context.Context, branch string) (string, error) {
			assert.Equal(t, "master", branch)
			return "sha", nil
		},
	}

	res := execTmpl(t, eng, `{{ lastCommit "master" }}`, nil)
	assert.Equal(t, "sha", res)
}

func TestTemplateFuncs_tags(t *testing.T) {
	eng := &RepositoryMock{
		ListTagsFunc: func(ctx context.Context) ([]Tag, error) {
			return []Tag{{Name: "v0.1.0"}, {Name: "v0.2.0"}}, nil
		},
	}

	res := execTmpl(t, eng, `{{ tags }}`, nil)
	assert.Equal(t, fmt.Sprintf("%v", []string{"v0.2.0", "v0.1.0"}), res)
}

func TestTemplateFuncs_headed(t *testing.T) {
	res := execTmpl(t, nil, `{{ headed .List }}`, struct{ List []string }{List: []string{"v0.1.0"}})
	assert.Equal(t, fmt.Sprintf("%v", []string{"HEAD", "v0.1.0"}), res)
}

func TestTemplateFuncs_prTitles(t *testing.T) {
	res := execTmpl(t, nil, `{{ prTitles .List }}`, struct{ List []PullRequest }{
		List: []PullRequest{
			{Title: "title1"},
			{Title: "title2"},
			{Title: "title3"},
		},
	})
	assert.Equal(t, fmt.Sprintf("%v", []string{"title1", "title2", "title3"}), res)
}

func execTmpl(t *testing.T, eng Repository, expr string, data any) string {
	fns, err := (&TemplateFuncs{Repository: eng}).Funcs(context.Background())
	require.NoError(t, err)

	tmpl, err := template.New("").Funcs(fns).Parse(expr)
	require.NoError(t, err)

	buf := &bytes.Buffer{}
	require.NoError(t, tmpl.Execute(buf, data))

	return buf.String()
}

func TestTemplateFuncs_String(t *testing.T) {
	assert.Equal(t, "git", (&TemplateFuncs{}).String())
}
