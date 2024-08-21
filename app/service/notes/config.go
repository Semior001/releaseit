package notes

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config describes the configuration of the changelog builder.
type Config struct {
	Categories []CategoryConfig `yaml:"categories"` // categories to parse in pull requests

	// field, by which pull requests must be sorted, in format +|-field
	// currently supported fields: number, author, title, closed
	SortField string `yaml:"sort_field"`

	Template     string   `yaml:"template"`      // template for a changelog.
	UnusedTitle  string   `yaml:"unused_title"`  // if set, the unused category will be built under this title at the, end of the changelog
	IgnoreLabels []string `yaml:"ignore_labels"` // labels for pull requests, which won't be in release notes
}

// CategoryConfig describes the category configuration.
type CategoryConfig struct {
	Title  string   `yaml:"title"`
	Labels []string `yaml:"labels"`

	Branch        string `yaml:"branch"`         // regexp to match source branch name
	CommitMessage string `yaml:"commit_message"` // regexp to match commit message

	// next fields are used internally
	BranchRe    *regexp.Regexp `yaml:"-"`
	CommitMsgRe *regexp.Regexp `yaml:"-"`
}

func (c *Config) validate() error {
	if len(c.Categories) == 0 {
		return errors.New("categories are empty")
	}

	if strings.TrimSpace(c.Template) == "" {
		return errors.New("template is empty")
	}

	for idx, category := range c.Categories {
		if category.Branch != "" {
			re, err := regexp.Compile(category.Branch)
			if err != nil {
				return fmt.Errorf("invalid regexp for branch: %w", err)
			}
			c.Categories[idx].BranchRe = re
		}

		if category.CommitMessage != "" {
			re, err := regexp.Compile(category.CommitMessage)
			if err != nil {
				return fmt.Errorf("invalid regexp for commit message: %w", err)
			}
			c.Categories[idx].CommitMsgRe = re
		}
	}

	return nil
}

const defaultTemplate = `Version {{.To}}
{{if (eq .Total 0)}}- No changes{{end}}{{range .Categories}}{{.Title}}
{{range .PRs}}- {{.Title}} (#{{.Number}}) by @{{.Author}}{{end}}
{{end}}`

func (c *Config) defaults() {
	if c.Template == "" {
		c.Template = defaultTemplate
	}
}

// ConfigFromFile reads the configuration from the file.
func ConfigFromFile(path string) (Config, error) {
	bts, err := os.ReadFile(path) //nolint:gosec // we don't need to check permissions here
	if err != nil {
		return Config{}, fmt.Errorf("open file: %w", err)
	}

	var res Config

	if err = yaml.Unmarshal(bts, &res); err != nil {
		return Config{}, fmt.Errorf("parse yaml: %w", err)
	}

	if err = res.validate(); err != nil {
		return Config{}, fmt.Errorf("config is invalid: %w", err)
	}

	res.defaults()

	return res, nil
}
