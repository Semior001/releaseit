package git

import (
	"context"
	"fmt"
	"github.com/samber/lo"
	"text/template"
)

// TemplateFuncs is a wrapper for Repository that provides functions for use in templates.
type TemplateFuncs struct{ Repository }

// String returns the name of the template addon.
func (g *TemplateFuncs) String() string { return "git" }

// Funcs returns a map of functions for use in templates.
func (g *TemplateFuncs) Funcs(ctx context.Context) (template.FuncMap, error) {
	return template.FuncMap{
		"previousTag": g.previousTag(ctx),
		"lastCommit":  g.lastCommit(ctx),
		"tags":        g.tags(ctx),
		"headed":      func(vals []string) []string { return append([]string{"HEAD"}, vals...) },
	}, nil
}

func (g *TemplateFuncs) lastCommit(ctx context.Context) func(branch string) (string, error) {
	return func(branch string) (string, error) {
		return g.Repository.GetLastCommitOfBranch(ctx, branch)
	}
}

func (g *TemplateFuncs) tags(ctx context.Context) func() ([]string, error) {
	return func() ([]string, error) {
		tags, err := g.Repository.ListTags(ctx)
		if err != nil {
			return nil, fmt.Errorf("list tags: %w", err)
		}

		// revert the order of tags
		for i, j := 0, len(tags)-1; i < j; i, j = i+1, j-1 {
			tags[i], tags[j] = tags[j], tags[i]
		}

		return lo.Map(tags, func(tag Tag, _ int) string { return tag.Name }), nil
	}
}

func (g *TemplateFuncs) previousTag(ctx context.Context) func(commitAlias string, tags []string) (string, error) {
	return func(commitAlias string, tagNames []string) (string, error) {
		tags, err := g.Repository.ListTags(ctx)
		if err != nil {
			return "", fmt.Errorf("list tags: %w", err)
		}

		tags = lo.Filter(tags, func(tag Tag, _ int) bool { return lo.Contains(tagNames, tag.Name) })

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
			comp, err := g.Repository.Compare(ctx, tag.Name, commitAlias)
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
