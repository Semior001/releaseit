// Package cmd defines commands of the application.
package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Semior001/releaseit/app/git/engine"
	"github.com/Semior001/releaseit/app/notify"
)

// EngineGroup defines parameters for the engine.
type EngineGroup struct {
	Type   string      `long:"type" env:"TYPE" choice:"github" choice:"gitlab" description:"type of the repository engine" required:"true"`
	Github GithubGroup `group:"github" namespace:"github" env-namespace:"GITHUB"`
	Gitlab GitlabGroup `group:"gitlab" namespace:"gitlab" env-namespace:"GITLAB"`
}

// Build builds the engine.
func (r EngineGroup) Build() (engine.Interface, error) {
	switch r.Type {
	case "github":
		return engine.NewGithub(
			r.Github.Repo.Owner,
			r.Github.Repo.Name,
			r.Github.BasicAuth.Username,
			r.Github.BasicAuth.Password,
			http.Client{Timeout: r.Github.Timeout},
		)
	case "gitlab":
		return engine.NewGitlab(
			r.Gitlab.Token,
			r.Gitlab.BaseURL,
			r.Gitlab.ProjectID,
			http.Client{Timeout: r.Gitlab.Timeout},
		)
	}
	return nil, fmt.Errorf("unsupported repository engine type %s", r.Type)
}

// GithubGroup defines parameters to connect to the github repository.
type GithubGroup struct {
	Repo struct {
		Owner string `long:"owner" env:"OWNER" description:"owner of the repository"`
		Name  string `long:"name" env:"NAME" description:"name of the repository"`
	} `group:"repo" namespace:"repo" env-namespace:"REPO"`
	BasicAuth struct {
		Username string `long:"username" env:"USERNAME" description:"username for basic auth"`
		Password string `long:"password" env:"PASSWORD" description:"password for basic auth"`
	} `group:"basic_auth" namespace:"basic_auth" env-namespace:"BASIC_AUTH"`
	Timeout time.Duration `long:"timeout" env:"TIMEOUT" description:"timeout for http requests" default:"5s"`
}

// GitlabGroup defines parameters to connect to the gitlab repository.
type GitlabGroup struct {
	Token     string        `long:"token" env:"TOKEN" description:"token to connect to the gitlab repository"`
	BaseURL   string        `long:"base_url" env:"BASE_URL" description:"base url of the gitlab instance"`
	ProjectID string        `long:"project_id" env:"PROJECT_ID" description:"project id of the repository"`
	Timeout   time.Duration `long:"timeout" env:"TIMEOUT" description:"timeout for http requests" default:"5s"`
}

// NotifyGroup defines parameters for the notifier.
type NotifyGroup struct {
	Telegram   TelegramGroup       `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`
	Github     GithubNotifierGroup `group:"github" namespace:"github" env-namespace:"GITHUB"`
	Mattermost MattermostGroup     `group:"mattermost-hook" namespace:"mattermost-hook" env-namespace:"MATTERMOST_HOOK"`
	Post       PostGroup           `group:"post" namespace:"post" env-namespace:"POST"`
	Stdout     bool                `long:"stdout" env:"STDOUT" description:"print release notes to stdout"`
}

// GithubNotifierGroup defines parameters to make release in the github.
type GithubNotifierGroup struct {
	GithubGroup
	ReleaseNameTemplate string `long:"release_name_tmpl" env:"RELEASE_NAME_TMPL" description:"template for release name"`
}

func (g GithubNotifierGroup) build() (notify.Destination, error) {
	return notify.NewGithub(notify.GithubParams{
		Owner:               g.Repo.Owner,
		Name:                g.Repo.Name,
		BasicAuthUsername:   g.BasicAuth.Username,
		BasicAuthPassword:   g.BasicAuth.Password,
		HTTPClient:          http.Client{Timeout: g.Timeout},
		ReleaseNameTmplText: g.ReleaseNameTemplate,
	})
}

// TelegramGroup defines parameters for telegram notifier.
type TelegramGroup struct {
	ChatID         string        `long:"chat_id" env:"CHAT_ID" description:"id of the chat, where the release notes will be sent"`
	Token          string        `long:"token" env:"TOKEN" description:"bot token"`
	WebPagePreview bool          `long:"web_page_preview" env:"WEB_PAGE_PREVIEW" description:"request telegram to preview for web links"`
	Timeout        time.Duration `long:"timeout" env:"TIMEOUT" description:"timeout for http requests" default:"5s"`
}

func (g TelegramGroup) build() (notify.Destination, error) {
	return notify.NewTelegram(notify.TelegramParams{
		Log:                   log.Default(),
		ChatID:                g.ChatID,
		Client:                http.Client{Timeout: g.Timeout},
		Token:                 g.Token,
		DisableWebPagePreview: !g.WebPagePreview,
	}), nil
}

// MattermostGroup defines parameters for mattermost hook notifier.
type MattermostGroup struct {
	BaseURL string        `long:"base_url" env:"BASE_URL" description:"base url of the mattermost server"`
	ID      string        `long:"id" env:"ID" description:"id of the hook, where the release notes will be sent"`
	Timeout time.Duration `long:"timeout" env:"TIMEOUT" description:"timeout for http requests" default:"5s"`
}

func (g MattermostGroup) build() (notify.Destination, error) {
	return notify.NewMattermost(
		http.Client{Timeout: g.Timeout},
		g.BaseURL,
		g.ID,
	), nil
}

// PostGroup defines parameters for post notifier.
type PostGroup struct {
	URL     string        `long:"url" env:"URL" description:"url to send the release notes"`
	Timeout time.Duration `long:"timeout" env:"TIMEOUT" description:"timeout for http requests" default:"5s"`
}

func (g PostGroup) build() (notify.Destination, error) {
	return &notify.Post{
		URL:    g.URL,
		Client: &http.Client{Timeout: g.Timeout},
	}, nil
}

// Build builds the notifier.
func (r *NotifyGroup) Build() (destinations notify.Destinations, err error) {
	if r.Stdout {
		destinations = append(destinations, &notify.WriterNotifier{Writer: os.Stdout, Name: "stdout"})
	}

	for _, d := range []struct {
		name  string
		empty bool
		build func() (notify.Destination, error)
	}{
		{name: "telegram", empty: r.Telegram.empty(), build: r.Telegram.build},
		{name: "github", empty: r.Github.empty(), build: r.Github.build},
		{name: "mattermost-hook", empty: r.Mattermost.empty(), build: r.Mattermost.build},
		{name: "post", empty: r.Post.empty(), build: r.Post.build},
	} {
		if d.empty {
			continue
		}

		dest, err := d.build()
		if err != nil {
			return nil, fmt.Errorf("failed to build %s notifier: %w", d.name, err)
		}
		destinations = append(destinations, dest)
	}

	log.Printf("[INFO] initialized %d notifiers: %s", len(destinations), destinations.String())

	return destinations, nil
}

func (g PostGroup) empty() bool       { return g.URL == "" }
func (g MattermostGroup) empty() bool { return g.BaseURL == "" || g.ID == "" }
func (g TelegramGroup) empty() bool   { return g.ChatID == "" || g.Token == "" }
func (g GithubNotifierGroup) empty() bool {
	return g.ReleaseNameTemplate == "" || g.Repo.Owner == "" || g.Repo.Name == ""
}
