// Package config contains the definition of the configuration file.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config defines configuration for the release notes builder.
type Config struct {
	// categories to parse in pull requests
	Categories []struct {
		Title  string   `yaml:"title"`
		Labels []string `yaml:"labels"`
	} `yaml:"categories"`
	// labels for pull requests, which won't be in release notes
	IgnoreLabels []string `yaml:"ignore_labels"`
	// field, by which pull requests must be sorted, in format +|-field
	// currently supported fields: number, author, title, closed
	SortField string `yaml:"sort_field"`
	// template for a changelog.
	Template string `yaml:"template"`
	// template for release with no changes
	EmptyTemplate string `yaml:"empty_template"`
	// if set, the unused category will be built under this title at the
	// end of the changelog
	UnusedTitle string `yaml:"unused_title"`
}

// Validate validates the configuration.
func (c Config) Validate() error {
	if len(c.Categories) == 0 {
		return errors.New("categories are empty")
	}

	if strings.TrimSpace(c.Template) == "" {
		return errors.New("template is empty")
	}

	if strings.TrimSpace(c.EmptyTemplate) == "" {
		return errors.New("template for empty changelog is empty")
	}

	return nil
}

// Read parses and validates the configuration
func Read(path string) (Config, error) {
	bytes, err := os.ReadFile(path) //nolint:gosec // we don't need to check permissions here
	if err != nil {
		return Config{}, fmt.Errorf("open file: %w", err)
	}

	var res Config

	if err = yaml.Unmarshal(bytes, &res); err != nil {
		return Config{}, fmt.Errorf("parse yaml: %w", err)
	}

	// validating config
	if err = res.Validate(); err != nil {
		return Config{}, fmt.Errorf("config is invalid: %w", err)
	}

	return res, nil
}
