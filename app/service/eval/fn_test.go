package eval

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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