// Package eval provides a common evaluator for templated expressions, that
// may make requests to remote services.
package eval

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/samber/lo"
)

// Evaluator is a service that provides common functions for most of the
// consumers for evaluating go template expressions.
type Evaluator struct {
	Addon Addon
}

// Validate validates the expression.
func (s *Evaluator) Validate(expr string) error {
	fm, err := s.funcs(context.Background())
	if err != nil {
		return fmt.Errorf("build funcs: %w", err)
	}

	if _, err = template.New("").Funcs(fm).Parse(expr); err != nil {
		return parseError{err: err}
	}

	return nil
}

// Evaluate evaluates the provided expression with the given data.
func (s *Evaluator) Evaluate(ctx context.Context, expr string, data any) (string, error) {
	buf := &bytes.Buffer{}

	fm, err := s.funcs(ctx)
	if err != nil {
		return "", fmt.Errorf("build funcs: %w", err)
	}

	tmpl, err := template.New("").
		Funcs(fm).
		Parse(expr)
	if err != nil {
		return "", parseError{err: err}
	}

	if err = tmpl.Execute(buf, data); err != nil {
		return "", fmt.Errorf("execute expression: %w", err)
	}

	return buf.String(), nil
}

func (s *Evaluator) funcs(ctx context.Context) (template.FuncMap, error) {
	funcs := lo.Assign(
		lo.OmitByKeys(sprig.FuncMap(), []string{"env", "expandenv"}),
		template.FuncMap{
			"next":     next,
			"previous": previous,
			"filter":   filter,
			"strings":  stringsFromAnys,

			// constants
			"semver": func() string { return `^v?(\d+)\.(\d+)\.(\d+)$` },
		},
	)

	if s.Addon != nil {
		fm, err := s.Addon.Funcs(ctx)
		if err != nil {
			return nil, fmt.Errorf("build addon %s: %w", s.Addon, err)
		}

		funcs = lo.Assign(funcs, fm)
	}

	return funcs, nil
}

func filter(rx string, elems []string) (res []string, err error) {
	r, err := regexp.Compile(rx)
	if err != nil {
		return nil, fmt.Errorf("compile regexp: %w", err)
	}

	for _, e := range elems {
		if r.MatchString(e) {
			res = append(res, e)
		}
	}

	return res, nil
}

func next(elem string, elems []string) string {
	for idx, e := range elems {
		if e == elem {
			if idx+1 == len(elems) {
				return ""
			}

			return elems[idx+1]
		}
	}

	return ""
}

func previous(elem string, elems []string) string {
	for idx, e := range elems {
		if e == elem {
			if idx == 0 {
				return ""
			}

			return elems[idx-1]
		}
	}

	return ""
}

func stringsFromAnys(elems []interface{}) []string {
	out, _ := lo.FromAnySlice[string](elems)
	return out
}

type parseError struct{ err error }

func (e parseError) Error() string {
	return fmt.Sprintf("parse expression: %v", e.err)
}
