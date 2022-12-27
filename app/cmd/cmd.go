package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// GithubGroup defines parameters to connect to the github repository.
type GithubGroup struct {
	Repo struct {
		Owner string `long:"owner" env:"OWNER" description:"owner of the repository"`
		Name  string `long:"name" env:"NAME" description:"name of the repository"`
	} `group:"repo" namespace:"repo" env-namespace:"REPO"`
	BasicAuth BasicAuthGroup `group:"basic_auth" namespace:"basic_auth" env-namespace:"BASIC_AUTH"`
}

// GitlabGroup defines parameters to connect to the gitlab repository.
type GitlabGroup struct {
	Token     string `long:"token" env:"TOKEN" description:"token to connect to the gitlab repository"`
	BaseURL   string `long:"base_url" env:"BASE_URL" description:"base url of the gitlab instance"`
	ProjectID string `long:"project_id" env:"PROJECT_ID" description:"project id of the repository"`
}

// Empty returns true if the argument group is empty.
func (g GithubGroup) Empty() bool {
	return g.Repo.Owner == "" || g.Repo.Name == ""
}

// GithubNotifierGroup defines parameters to make release in the github.
type GithubNotifierGroup struct {
	GithubGroup
	ReleaseNameTemplate string `long:"release_name_tmpl" env:"RELEASE_NAME_TMPL" description:"template for release name"`
	ConfLocation        string `long:"conf_location" env:"CONF_LOCATION" description:"location to the config file"`
}

// Empty returns true if the argument group is empty.
func (g GithubNotifierGroup) Empty() bool {
	return g.ReleaseNameTemplate == "" || g.GithubGroup.Empty()
}

// BasicAuthGroup defines parameters for basic authentication.
type BasicAuthGroup struct {
	Username string `long:"username" env:"USERNAME" description:"username for basic auth"`
	Password string `long:"password" env:"PASSWORD" description:"password for basic auth"`
}

// StdoutGroup defines parameters for printing release notes to stdout.
type StdoutGroup struct {
	ConfLocation string `long:"conf_location" env:"CONF_LOCATION" description:"location to the config file"`
}

// Empty returns true if stdout group is empty.
func (g StdoutGroup) Empty() bool {
	return g.ConfLocation == ""
}

// TelegramGroup defines parameters for telegram notifier.
type TelegramGroup struct {
	ChatID         string `long:"chat_id" env:"CHAT_ID" description:"id of the chat, where the release notes will be sent"`
	Token          string `long:"token" env:"TOKEN" description:"bot token"`
	WebPagePreview bool   `long:"web_page_preview" env:"WEB_PAGE_PREVIEW" description:"request telegram to preview for web links"`
	ConfLocation   string `long:"conf_location" env:"CONF_LOCATION" description:"location to the config file"`
}

// Empty returns true if the config group is not filled.
func (g TelegramGroup) Empty() bool {
	return g.ConfLocation == "" || g.ChatID == "" || g.Token == ""
}

type changelogBuilderCfg struct {
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

func (c changelogBuilderCfg) Validate() error {
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

func parseCfg(cfgPath string) (changelogBuilderCfg, error) {
	bytes, err := os.ReadFile(cfgPath)
	if err != nil {
		return changelogBuilderCfg{}, fmt.Errorf("open file: %w", err)
	}

	var res changelogBuilderCfg

	if err = yaml.Unmarshal(bytes, &res); err != nil {
		return changelogBuilderCfg{}, fmt.Errorf("parse yaml: %w", err)
	}

	// validating config
	if err = res.Validate(); err != nil {
		return changelogBuilderCfg{}, fmt.Errorf("config is invalid: %w", err)
	}

	return res, nil
}
