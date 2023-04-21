// Package eval provides a common evaluator for templated expressions, that
// may make requests to remote services.
package eval

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/git/engine"
	"github.com/samber/lo"
)

// Evaluator is a service that provides common functions for most of the
// consumers for evaluating go template expressions.
type Evaluator struct {
	Funcs  template.FuncMap
	Engine engine.Interface
}

// Evaluate evaluates the provided expression with the given data.
func (s *Evaluator) Evaluate(ctx context.Context, expr string, data any) (string, error) {
	buf := &bytes.Buffer{}

	tmpl, err := template.New("").
		Funcs(s.funcs(ctx)).
		Parse(expr)
	if err != nil {
		return "", parseError{err: err}
	}

	if err = tmpl.Execute(buf, data); err != nil {
		return "", fmt.Errorf("execute expression: %w", err)
	}

	return buf.String(), nil
}

func (s *Evaluator) funcs(ctx context.Context) template.FuncMap {
	fns := template.FuncMap{}
	fns["last_commit"] = s.lastCommit(ctx)
	fns["previous_tag"] = s.previousTag(ctx)
	fns["last_tag"] = s.lastTag(ctx)
	fns["tags"] = s.tags(ctx)
	fns = lo.Assign(fns,
		lo.OmitByKeys(sprig.FuncMap(), []string{"env", "expandenv"}),
		s.Funcs,
	)
	return fns
}

func (s *Evaluator) lastCommit(ctx context.Context) func(branch string) (string, error) {
	return func(branch string) (string, error) {
		return s.Engine.GetLastCommitOfBranch(ctx, branch)
	}
}

func (s *Evaluator) previousTag(ctx context.Context) func(commitAlias string) (string, error) {
	return func(commitAlias string) (string, error) {
		tags, err := s.Engine.ListTags(ctx)
		if err != nil {
			return "", fmt.Errorf("list tags: %w", err)
		}

		// if by any chance alias is a tag itself
		for idx, tag := range tags {
			if tag.Name == commitAlias || tag.Commit.SHA == commitAlias {
				if idx+1 == len(tags) {
					return "HEAD", nil
				}

				return tags[idx+1].Name, nil
			}
		}

		// otherwise, we find the closest tag
		for _, tag := range tags {
			comp, err := s.Engine.Compare(ctx, tag.Name, commitAlias)
			if err != nil {
				return "", fmt.Errorf("compare tag %s with commit %s: %w",
					tag.Commit.SHA, commitAlias, err)
			}

			if len(comp.Commits) > 0 {
				return tag.Name, nil
			}
		}

		return "HEAD", nil
	}
}

func (s *Evaluator) lastTag(ctx context.Context) func() (string, error) {
	return func() (string, error) {
		tags, err := s.Engine.ListTags(ctx)
		if err != nil {
			return "", fmt.Errorf("list tags: %w", err)
		}

		if len(tags) == 0 {
			return "HEAD", nil
		}

		return tags[0].Name, nil
	}
}

func (s *Evaluator) tags(ctx context.Context) func() ([]string, error) {
	return func() ([]string, error) {
		tags, err := s.Engine.ListTags(ctx)
		if err != nil {
			return nil, fmt.Errorf("list tags: %w", err)
		}

		return lo.Map(tags, func(tag git.Tag, _ int) string { return tag.Name }), nil
	}
}

type parseError struct{ err error }

func (e parseError) Error() string { return fmt.Sprintf("parse expression: %v", e.err) }
