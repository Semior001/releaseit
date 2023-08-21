package eval

import (
	"context"
	"fmt"
	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/git/engine"
	"github.com/samber/lo"
	"text/template"
)

// Git is an addon for evaluating git-related functions in templates.
type Git struct {
	Engine engine.Interface
}

// String returns the name of the template addon.
func (g *Git) String() string { return "git" }

// Funcs returns a map of functions for use in templates.
func (g *Git) Funcs(ctx context.Context) (template.FuncMap, error) {
	return template.FuncMap{
		"previousTag": g.previousTag(ctx),
		"lastCommit":  g.lastCommit(ctx),
		"tags":        g.tags(ctx),
		"headed":      g.headed,
		"prTitles":    g.prTitles,
	}, nil
}

func (g *Git) prTitles(prs []git.PullRequest) []string {
	titles := make([]string, len(prs))
	for i, pr := range prs {
		titles[i] = pr.Title
	}
	return titles
}

func (g *Git) headed(vals []string) []string {
	return append([]string{"HEAD"}, vals...)
}

func (g *Git) lastCommit(ctx context.Context) func(branch string) (string, error) {
	return func(branch string) (string, error) {
		return g.Engine.GetLastCommitOfBranch(ctx, branch)
	}
}

func (g *Git) tags(ctx context.Context) func() ([]string, error) {
	return func() ([]string, error) {
		tags, err := g.Engine.ListTags(ctx)
		if err != nil {
			return nil, fmt.Errorf("list tags: %w", err)
		}

		// revert the order of tags
		for i, j := 0, len(tags)-1; i < j; i, j = i+1, j-1 {
			tags[i], tags[j] = tags[j], tags[i]
		}

		return lo.Map(tags, func(tag git.Tag, _ int) string { return tag.Name }), nil
	}
}

func (g *Git) previousTag(ctx context.Context) func(commitAlias string, tags []string) (string, error) {
	return func(commitAlias string, tagNames []string) (string, error) {
		tags, err := g.Engine.ListTags(ctx)
		if err != nil {
			return "", fmt.Errorf("list tags: %w", err)
		}

		tags = lo.Filter(tags, func(tag git.Tag, _ int) bool { return lo.Contains(tagNames, tag.Name) })

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
			comp, err := g.Engine.Compare(ctx, tag.Name, commitAlias)
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
