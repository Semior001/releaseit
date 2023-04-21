package eval

import (
	"context"
	"fmt"
	"testing"
	"text/template"

	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/git/engine"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluator_EvaluateLastTag(t *testing.T) {
	svc := &Evaluator{
		Engine: &engine.InterfaceMock{
			ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
				return []git.Tag{{Name: "v0.2.0"}, {Name: "v0.1.0"}}, nil
			},
		},
	}

	res, err := svc.Evaluate(context.Background(), `{{ last_tag }}`, nil)
	require.NoError(t, err)
	assert.Equal(t, "v0.2.0", res)
}

func TestEvaluator_EvaluateLastCommit(t *testing.T) {
	svc := &Evaluator{Engine: &engine.InterfaceMock{
		GetLastCommitOfBranchFunc: func(ctx context.Context, branch string) (string, error) {
			assert.Equal(t, "master", branch)
			return "sha", nil
		},
	}}

	res, err := svc.Evaluate(context.Background(), `{{ last_commit "master" }}`, nil)
	require.NoError(t, err)
	assert.Equal(t, "sha", res)
}

func TestEvaluator_EvaluateTags(t *testing.T) {
	svc := &Evaluator{Engine: &engine.InterfaceMock{
		ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
			return []git.Tag{{Name: "v0.1.0"}, {Name: "v0.2.0"}}, nil
		},
	}}

	res, err := svc.Evaluate(context.Background(), "{{ tags }}", nil)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%v", []string{"v0.1.0", "v0.2.0"}), res)
}

func TestEvaluator_EvaluatePreviousTag(t *testing.T) {
	t.Run("last tag", func(t *testing.T) {
		svc := &Evaluator{Engine: &engine.InterfaceMock{
			ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
				return []git.Tag{{Name: "v0.2.0"}, {Name: "v0.1.0"}}, nil
			},
		}}

		res, err := svc.Evaluate(context.Background(), `{{ previous_tag "v0.2.0" }}`, nil)
		require.NoError(t, err)
		assert.Equal(t, "v0.1.0", res)
	})

	t.Run("first tag", func(t *testing.T) {
		svc := &Evaluator{Engine: &engine.InterfaceMock{
			ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
				return []git.Tag{{Name: "v0.1.0"}}, nil
			},
		}}

		res, err := svc.Evaluate(context.Background(), `{{ previous_tag "v0.1.0" }}`, nil)
		require.NoError(t, err)
		assert.Equal(t, "HEAD", res)
	})

	t.Run("tag's commit SHA", func(t *testing.T) {
		svc := &Evaluator{Engine: &engine.InterfaceMock{
			ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
				return []git.Tag{
					{Name: "v0.2.0", Commit: git.Commit{SHA: "sha"}},
					{Name: "v0.1.0"},
				}, nil
			},
		}}

		res, err := svc.Evaluate(context.Background(), `{{ previous_tag "sha" }}`, nil)
		require.NoError(t, err)
		assert.Equal(t, "v0.1.0", res)
	})

	t.Run("arbitrary commit SHA", func(t *testing.T) {
		svc := &Evaluator{Engine: &engine.InterfaceMock{
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

		res, err := svc.Evaluate(context.Background(), `{{ previous_tag "sha" }}`, nil)
		require.NoError(t, err)
		assert.Equal(t, "v0.1.0", res)
	})

	t.Run("nothing found", func(t *testing.T) {
		svc := &Evaluator{Engine: &engine.InterfaceMock{
			ListTagsFunc: func(ctx context.Context) ([]git.Tag, error) {
				return []git.Tag{{Name: "v0.1.0"}}, nil
			},
			CompareFunc: func(ctx context.Context, from, to string) (git.CommitsComparison, error) {
				assert.Equal(t, "sha", to)
				assert.Equal(t, "v0.1.0", from)
				return git.CommitsComparison{}, nil
			},
		}}

		res, err := svc.Evaluate(context.Background(), `{{ previous_tag "sha" }}`, nil)
		require.NoError(t, err)
		assert.Equal(t, "HEAD", res)
	})
}

func TestEvaluator_EvaluateCustomFunction(t *testing.T) {
	svc := &Evaluator{
		Funcs: template.FuncMap{
			"custom_func": func() string { return "some custom output" },
		},
	}

	res, err := svc.Evaluate(context.Background(), `{{ custom_func }}`, nil)
	require.NoError(t, err)
	assert.Equal(t, "some custom output", res)
}

func TestEvaluator_EvaluateSprigFuncs(t *testing.T) {
	t.Run("env and expandenv should be omitted", func(t *testing.T) {
		svc := &Evaluator{}

		res, err := svc.Evaluate(context.Background(), `{{ env "SOME_VAR" }}`, nil)
		require.ErrorAs(t, err, &parseError{})
		assert.Equal(t, "", res)

		res, err = svc.Evaluate(context.Background(), `{{ expandenv "SOME_VAR" }}`, nil)
		require.ErrorAs(t, err, &parseError{})
		assert.Equal(t, "", res)
	})
}
