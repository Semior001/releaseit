package eval

import (
	"context"
	"fmt"
	"strings"
	"text/template"
)

//go:generate rm -f mock_addon.go
//go:generate moq -out mock_addon.go . Addon

// Addon extends the evaluator with additional functions.
type Addon interface {
	fmt.Stringer
	Funcs(ctx context.Context) (template.FuncMap, error)
}

// MultiAddon is a list of addons.
type MultiAddon []Addon

// Name returns names of all underlying addons.
func (m MultiAddon) Name() string {
	dests := make([]string, len(m))
	for i, dest := range m {
		dests[i] = dest.String()
	}
	return fmt.Sprintf("[%s]", strings.Join(dests, ", "))
}

// Funcs returns a merged func map of all underlying addons.
func (m MultiAddon) Funcs(ctx context.Context) (template.FuncMap, error) {
	type fnWithAddr struct {
		AddonName string
		Func      interface{}
	}

	funcs := map[string]fnWithAddr{}
	for _, addon := range m {
		f, err := addon.Funcs(ctx)
		if err != nil {
			return nil, fmt.Errorf("addon %s: %w", addon, err)
		}
		for name, fn := range f {
			if faddr, ok := funcs[name]; ok {
				return nil, fmt.Errorf("addon %s: function %s already defined by addon %s",
					addon, name, faddr.AddonName)
			}
			funcs[name] = fnWithAddr{AddonName: addon.String(), Func: fn}
		}
	}

	res := make(template.FuncMap, len(funcs))
	for name, fn := range funcs {
		res[name] = fn.Func
	}

	return res, nil
}
