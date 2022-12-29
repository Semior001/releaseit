package notes

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type config struct {
	// categories to parse in pull requests
	Categories []categoryConfig `yaml:"categories"`
	// labels for pull requests, which won't be in release notes
	IgnoreLabels []string `yaml:"ignore_labels"`
	// field, by which pull requests must be sorted, in format +|-field
	// currently supported fields: number, author, title, closed
	SortField string `yaml:"sort_field"`
	// template for a changelog.
	Template string `yaml:"template"`
	// if set, the unused category will be built under this title at the
	// end of the changelog
	UnusedTitle string `yaml:"unused_title"`
}

type categoryConfig struct {
	Title  string   `yaml:"title"`
	Labels []string `yaml:"labels"`

	// regexp to match branch name
	Branch   string         `yaml:"branch"`
	branchRe *regexp.Regexp `yaml:"-"`
}

func (c *config) validate() error {
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
			c.Categories[idx].branchRe = re
		}
	}

	return nil
}

func readCfg(path string) (config, error) {
	bts, err := os.ReadFile(path) //nolint:gosec // we don't need to check permissions here
	if err != nil {
		return config{}, fmt.Errorf("open file: %w", err)
	}

	var res config

	if err = yaml.Unmarshal(bts, &res); err != nil {
		return config{}, fmt.Errorf("parse yaml: %w", err)
	}

	if err = res.validate(); err != nil {
		return config{}, fmt.Errorf("config is invalid: %w", err)
	}

	return res, nil
}
