package eval

import (
	"context"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"text/template"
)

func TestEvaluator_Evaluate(t *testing.T) {
	t.Run("func from addon", func(t *testing.T) {
		svc := &Evaluator{
			Addon: &AddonMock{FuncsFunc: func(ctx context.Context) (template.FuncMap, error) {
				return template.FuncMap{"foo": func() string { return "bar" }}, nil
			}},
		}

		res, err := svc.Evaluate(context.Background(), `{{ foo }}`, nil)
		require.NoError(t, err)
		assert.Equal(t, "bar", res)
	})

	t.Run("env and expandenv should be omitted", func(t *testing.T) {
		svc := &Evaluator{}

		res, err := svc.Evaluate(context.Background(), `{{ env "SOME_VAR" }}`, nil)
		assert.EqualError(t, err, "parse expression: template: :1: function \"env\" not defined")
		require.ErrorAs(t, err, &parseError{})
		assert.Equal(t, "", res)

		res, err = svc.Evaluate(context.Background(), `{{ expandenv "SOME_VAR" }}`, nil)
		require.ErrorAs(t, err, &parseError{})
		assert.Equal(t, "", res)
	})
}

func TestEvaluator_EvaluateNext(t *testing.T) {
	t.Run("in the middle", func(t *testing.T) {
		svc := &Evaluator{}

		res, err := svc.Evaluate(context.Background(),
			`{{ next "2" .List }}`,
			struct{ List []string }{
				List: []string{"1", "2", "3"},
			},
		)
		require.NoError(t, err)

		assert.Equal(t, "3", res)
	})

	t.Run("at the end", func(t *testing.T) {
		svc := &Evaluator{}

		res, err := svc.Evaluate(context.Background(),
			`{{ next "3" .List }}`,
			struct{ List []string }{
				List: []string{"1", "2", "3"},
			},
		)
		require.NoError(t, err)

		assert.Equal(t, "", res)
	})

	t.Run("not found", func(t *testing.T) {
		svc := &Evaluator{}

		res, err := svc.Evaluate(context.Background(),
			`{{ next "4" .List }}`,
			struct{ List []string }{
				List: []string{"1", "2", "3"},
			},
		)
		require.NoError(t, err)

		assert.Equal(t, "", res)
	})
}

func TestEvaluator_EvaluatePrevious(t *testing.T) {
	t.Run("in the middle", func(t *testing.T) {
		svc := &Evaluator{}

		res, err := svc.Evaluate(context.Background(),
			`{{ previous "2" .List }}`,
			struct{ List []string }{
				List: []string{"1", "2", "3"},
			},
		)
		require.NoError(t, err)

		assert.Equal(t, "1", res)
	})

	t.Run("at the beginning", func(t *testing.T) {
		svc := &Evaluator{}

		res, err := svc.Evaluate(context.Background(),
			`{{ previous "1" .List }}`,
			struct{ List []string }{
				List: []string{"1", "2", "3"},
			},
		)
		require.NoError(t, err)

		assert.Equal(t, "", res)
	})

	t.Run("not found", func(t *testing.T) {
		svc := &Evaluator{}

		res, err := svc.Evaluate(context.Background(),
			`{{ previous "4" .List }}`,
			struct{ List []string }{
				List: []string{"1", "2", "3"},
			},
		)
		require.NoError(t, err)

		assert.Equal(t, "", res)
	})
}

func TestEvaluator_FilterSemver(t *testing.T) {
	svc := &Evaluator{}

	res, err := svc.Evaluate(context.Background(),
		`{{ filter semver .List }}`,
		struct{ List []string }{
			List: []string{"1", "3", "v1.2.3", "2", "v1.2.4", "v1.2.5", "4"},
		},
	)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%v", []string{"v1.2.3", "v1.2.4", "v1.2.5"}), res)
}

func TestEvaluator_Strings(t *testing.T) {
	svc := &Evaluator{}

	res, err := svc.Evaluate(context.Background(),
		`{{ strings .List }}`,
		struct{ List []interface{} }{
			List: []interface{}{"1", "3", "v1.2.3", "2", "v1.2.4", "v1.2.5", "4"},
		},
	)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%v", []string{"1", "3", "v1.2.3", "2", "v1.2.4", "v1.2.5", "4"}), res)
}

func TestEvaluator_Validate(t *testing.T) {
	t.Run("ok - without addons", func(t *testing.T) {
		svc := &Evaluator{}

		err := svc.Validate(`{{ .List }}`)
		require.NoError(t, err)
	})

	t.Run("fail - without addons", func(t *testing.T) {
		svc := &Evaluator{}

		err := svc.Validate(`{{ unknownFunc .List }}`)
		require.ErrorAs(t, err, &parseError{})
	})

	t.Run("fail - failed to build addon funcs", func(t *testing.T) {
		expectedErr := errors.New("some lousy error")
		svc := &Evaluator{
			Addon: &AddonMock{
				FuncsFunc: func(ctx context.Context) (template.FuncMap, error) {
					return nil, expectedErr
				},
			},
		}

		err := svc.Validate(`{{ .List }}`)
		require.ErrorIs(t, err, expectedErr)
	})

	t.Run("fail - with addons", func(t *testing.T) {
		svc := &Evaluator{
			Addon: &AddonMock{
				FuncsFunc: func(ctx context.Context) (template.FuncMap, error) {
					return template.FuncMap{"foo": func() string { return "bar" }}, nil
				},
			},
		}

		err := svc.Validate(`{{ foo }}{{ .List }}{{ unknownFunc }}`)
		require.ErrorAs(t, err, &parseError{})
	})
}
