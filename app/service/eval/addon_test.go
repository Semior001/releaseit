package eval

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"text/template"
)

func TestMultiAddon_String(t *testing.T) {
	assert.Equal(t, "[mock, mock1]", MultiAddon{
		&AddonMock{StringFunc: func() string { return "mock" }},
		&AddonMock{StringFunc: func() string { return "mock1" }},
	}.String())
}

func TestMultiAddon_Funcs(t *testing.T) {
	t.Run("no addons", func(t *testing.T) {
		fn, err := MultiAddon{}.Funcs(nil)
		assert.NoError(t, err)
		assert.Empty(t, fn)
	})

	t.Run("one addon", func(t *testing.T) {
		fn, err := MultiAddon{
			&AddonMock{
				StringFunc: func() string { return "mock" },
				FuncsFunc: func(ctx context.Context) (template.FuncMap, error) {
					return map[string]interface{}{"fn": func() {}}, nil
				},
			},
		}.Funcs(nil)
		assert.NoError(t, err)
		assert.Len(t, fn, 1)
	})

	t.Run("addon conflict", func(t *testing.T) {
		fn, err := MultiAddon{
			&AddonMock{
				StringFunc: func() string { return "mock" },
				FuncsFunc: func(ctx context.Context) (template.FuncMap, error) {
					return map[string]interface{}{"fn": func() {}}, nil
				},
			},
			&AddonMock{
				StringFunc: func() string { return "mock1" },
				FuncsFunc: func(ctx context.Context) (template.FuncMap, error) {
					return map[string]interface{}{"fn": func() {}}, nil
				},
			},
		}.Funcs(nil)
		assert.ErrorContainsf(t, err, "addon mock1: function fn already defined by addon mock", "")
		assert.Nil(t, fn)
	})

	t.Run("addon returned error", func(t *testing.T) {
		expectedErr := errors.New("some error")
		fn, err := MultiAddon{
			&AddonMock{
				StringFunc: func() string { return "mock" },
				FuncsFunc: func(ctx context.Context) (template.FuncMap, error) {
					return nil, expectedErr
				},
			},
		}.Funcs(nil)
		assert.ErrorIs(t, err, expectedErr)
		assert.Nil(t, fn)
	})

	t.Run("multiple addons without collisions", func(t *testing.T) {
		fn, err := MultiAddon{
			&AddonMock{
				StringFunc: func() string { return "mock" },
				FuncsFunc: func(ctx context.Context) (template.FuncMap, error) {
					return map[string]interface{}{"fn": func() {}}, nil
				},
			},
			&AddonMock{
				StringFunc: func() string { return "mock1" },
				FuncsFunc: func(ctx context.Context) (template.FuncMap, error) {
					return map[string]interface{}{"fn1": func() {}}, nil
				},
			},
		}.Funcs(nil)
		assert.NoError(t, err)
		assert.Len(t, fn, 2)
	})
}
