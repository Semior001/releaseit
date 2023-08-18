package eval

import (
	"context"
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
		require.ErrorAs(t, err, &parseError{})
		assert.Equal(t, "", res)

		res, err = svc.Evaluate(context.Background(), `{{ expandenv "SOME_VAR" }}`, nil)
		require.ErrorAs(t, err, &parseError{})
		assert.Equal(t, "", res)
	})
}
