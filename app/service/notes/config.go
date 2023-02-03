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
	// categories to parse in pull requests
	Categories []CategoryConfig `yaml:"categories"`

	// field, by which pull requests must be sorted, in format +|-field
	// currently supported fields: number, author, title, closed
	SortField string `yaml:"sort_field"`
	// template for a changelog.
	Template string `yaml:"template"`

	// if set, the unused category will be built under this title at the
	// end of the changelog
	UnusedTitle string `yaml:"unused_title"`
	// labels for pull requests, which won't be in release notes
	IgnoreLabels []string `yaml:"ignore_labels"`
	// regexp for pull request branches, which won't be in release notes
	IgnoreBranch string `yaml:"ignore_branch"`
	// compiled regexp, used internally
	IgnoreBranchRe *regexp.Regexp `yaml:"-"`
}

// CategoryConfig describes the category configuration.
type CategoryConfig struct {
	Title  string   `yaml:"title"`
	Labels []string `yaml:"labels"`

	// regexp to match source branch name
	Branch string `yaml:"branch"`

	// compiled branch regexp, used internally
	BranchRe *regexp.Regexp `yaml:"-"`
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
	}

	return nil
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

	return res, nil
}
