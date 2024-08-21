package notes

import (
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_validate(t *testing.T) {
	t.Run("empty categories", func(t *testing.T) {
		cfg := Config{}
		assert.ErrorContains(t, cfg.validate(), "categories are empty")
	})

	t.Run("empty template", func(t *testing.T) {
		cfg := Config{Categories: []CategoryConfig{{}}}
		assert.ErrorContains(t, cfg.validate(), "template is empty")
	})

	t.Run("invalid regexp", func(t *testing.T) {
		cfg := Config{Categories: []CategoryConfig{{Branch: `[\]`}}, Template: "test"}
		assert.ErrorContains(t, cfg.validate(), "invalid regexp for branch")
	})
}

const testCfg = `categories:
  - title: "**🚀 Features**"
    branch: "^(feat|feature)/"
  - title: "**🐛 Fixes**"
    branch: "^fix/"
  - title: "**🔧 Maintenance**"
    branch: "^chore/"
unused_title: "**❓ Unlabeled**"
ignore_labels: ["ignore"]
sort_field: "+closed"
template: |
  Version {{.To}}
  {{if (eq .Total 0)}}- No changes{{end}}{{range .Categories}}{{.Title}}
  {{range .PRs}}- {{.Title}} (#{{.Number}}) by @{{.Author}}{{end}}
  {{end}}`

func TestConfigFromFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "config.yaml")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	_, err = f.WriteString(testCfg)
	require.NoError(t, err)

	cfg, err := ConfigFromFile(f.Name())
	require.NoError(t, err)

	assert.Equal(t, Config{
		Categories: []CategoryConfig{
			{Title: "**🚀 Features**", Branch: "^(feat|feature)/", BranchRe: regexp.MustCompile("^(feat|feature)/")},
			{Title: "**🐛 Fixes**", Branch: "^fix/", BranchRe: regexp.MustCompile("^fix/")},
			{Title: "**🔧 Maintenance**", Branch: "^chore/", BranchRe: regexp.MustCompile("^chore/")},
		},
		SortField:    "+closed",
		Template:     defaultTemplate,
		UnusedTitle:  "**❓ Unlabeled**",
		IgnoreLabels: []string{"ignore"},
	}, cfg)
}
