// Package cmd defines commands of the application.
package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	gengine "github.com/Semior001/releaseit/app/git/engine"
	"github.com/Semior001/releaseit/app/notify"
	tengine "github.com/Semior001/releaseit/app/task/engine"
)

// EngineGroup defines parameters for the engine.
type EngineGroup struct {
	Type   string      `long:"type" env:"TYPE" choice:"github" choice:"gitlab" description:"type of the repository engine" required:"true"`
	Github GithubGroup `group:"github" namespace:"github" env-namespace:"GITHUB"`
	Gitlab GitlabGroup `group:"gitlab" namespace:"gitlab" env-namespace:"GITLAB"`
}

// Build builds the engine.
func (r EngineGroup) Build(ctx context.Context) (gengine.Interface, error) {
	switch r.Type {
	case "github":
		if err := r.Github.fill(); err != nil {
			return nil, err
		}

		return gengine.NewGithub(ctx, gengine.GithubParams{
			Owner:             r.Github.Repo.Owner,
			Name:              r.Github.Repo.Name,
			BasicAuthUsername: r.Github.BasicAuth.Username,
			BasicAuthPassword: r.Github.BasicAuth.Password,
			HTTPClient:        http.Client{Timeout: r.Github.Timeout},
		})
	case "gitlab":
		return gengine.NewGitlab(ctx,
			r.Gitlab.Token,
			r.Gitlab.BaseURL,
			r.Gitlab.ProjectID,
			http.Client{Timeout: r.Gitlab.Timeout},
		)
	}
	return nil, fmt.Errorf("unsupported repository engine type %s", r.Type)
}

// TaskGroup defines parameters for task service
type TaskGroup struct {
	Type string `long:"type" env:"TYPE" choice:"" choice:"jira" description:"type of the task tracker"`
	Jira Jira   `group:"jira" namespace:"jira" env-namespace:"JIRA"`
}

// Build builds the task service.
func (r TaskGroup) Build(ctx context.Context) (_ *tengine.Tracker, err error) {
	var eng tengine.Interface
	switch r.Type {
	case "jira":
		if eng, err = r.Jira.Build(ctx); err != nil {
			return nil, fmt.Errorf("build jira task tracker: %w", err)
		}
	case "":
		eng = &tengine.Unsupported{}
	default:
		return nil, fmt.Errorf("unsupported task tracker type %s", r.Type)
	}
	return &tengine.Tracker{Interface: eng}, nil
}

// Jira defines parameters for the jira task tracker.
type Jira struct {
	URL     string        `long:"url" env:"URL" description:"url of the jira instance"`
	Token   string        `long:"token" env:"TOKEN" description:"token to connect to the jira instance"`
	Timeout time.Duration `long:"timeout" env:"TIMEOUT" description:"timeout for http requests" default:"5s"`
}

// Build builds the jira engine.
func (r Jira) Build(ctx context.Context) (tengine.Interface, error) {
	return tengine.NewJira(ctx, tengine.JiraParams{
		URL:        r.URL,
		Token:      r.Token,
		HTTPClient: http.Client{Timeout: r.Timeout},
	})
}

// GithubGroup defines parameters to connect to the github repository.
type GithubGroup struct {
	Repo struct {
		FullName string `long:"full-name" env:"FULL_NAME" description:"full name of the repository (owner/name)"`
		Owner    string `long:"owner" env:"OWNER" description:"owner of the repository"`
		Name     string `long:"name" env:"NAME" description:"name of the repository"`
	} `group:"repo" namespace:"repo" env-namespace:"REPO"`
	BasicAuth struct {
		Username string `long:"username" env:"USERNAME" description:"username for basic auth"`
		Password string `long:"password" env:"PASSWORD" description:"password for basic auth"`
	} `group:"basic-auth" namespace:"basic-auth" env-namespace:"BASIC_AUTH"`
	Timeout time.Duration `long:"timeout" env:"TIMEOUT" description:"timeout for http requests" default:"5s"`
}

func (g *GithubGroup) fill() error {
	if g.Repo.FullName == "" || (g.Repo.Owner != "" && g.Repo.Name != "") {
		return nil
	}

	tokens := strings.Split(g.Repo.FullName, "/")
	if len(tokens) != 2 {
		return fmt.Errorf("invalid repository name %s", g.Repo.FullName)
	}
	g.Repo.Owner, g.Repo.Name = tokens[0], tokens[1]
	return nil
}

// GitlabGroup defines parameters to connect to the gitlab repository.
type GitlabGroup struct {
	Token     string        `long:"token" env:"TOKEN" description:"token to connect to the gitlab repository"`
	BaseURL   string        `long:"base-url" env:"BASE_URL" description:"base url of the gitlab instance"`
	ProjectID string        `long:"project-id" env:"PROJECT_ID" description:"project id of the repository"`
	Timeout   time.Duration `long:"timeout" env:"TIMEOUT" description:"timeout for http requests" default:"5s"`
}

// NotifyGroup defines parameters for the notifier.
type NotifyGroup struct {
	Telegram      TelegramGroup       `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`
	Github        GithubNotifierGroup `group:"github" namespace:"github" env-namespace:"GITHUB"`
	Mattermost    MattermostHookGroup `group:"mattermost-hook" namespace:"mattermost-hook" env-namespace:"MATTERMOST_HOOK"`
	MattermostBot MattermostBotGroup  `group:"mattermost-bot" namespace:"mattermost-bot" env-namespace:"MATTERMOST_BOT"`
	Post          PostGroup           `group:"post" namespace:"post" env-namespace:"POST"`
	Stdout        bool                `long:"stdout" env:"STDOUT" description:"print release notes to stdout"`
	Stderr        bool                `long:"stderr" env:"STDERR" description:"print release notes to stderr"`
}

// GithubNotifierGroup defines parameters to make release in the github.
type GithubNotifierGroup struct {
	GithubGroup
	ReleaseNameTemplate string            `long:"release-name-tmpl" env:"RELEASE_NAME_TMPL" description:"template for release name"`
	Tag                 string            `long:"tag" env:"TAG" description:"tag to specify release"`
	Extras              map[string]string `long:"extra" env:"EXTRA" description:"extra parameters to pass to the notifier"`
}

func (g GithubNotifierGroup) build() (notify.Destination, error) {
	if err := g.GithubGroup.fill(); err != nil {
		return nil, err
	}

	return notify.NewGithub(notify.GithubParams{
		Owner:               g.Repo.Owner,
		Name:                g.Repo.Name,
		BasicAuthUsername:   g.BasicAuth.Username,
		BasicAuthPassword:   g.BasicAuth.Password,
		HTTPClient:          http.Client{Timeout: g.Timeout},
		ReleaseNameTmplText: g.ReleaseNameTemplate,
		Tag:                 g.Tag,
		Extras:              g.Extras,
	})
}

// TelegramGroup defines parameters for telegram notifier.
type TelegramGroup struct {
	ChatID         string        `long:"chat-id" env:"CHAT_ID" description:"id of the chat, where the release notes will be sent"`
	Token          string        `long:"token" env:"TOKEN" description:"bot token"`
	WebPagePreview bool          `long:"web-page-preview" env:"WEB_PAGE_PREVIEW" description:"request telegram to preview for web links"`
	Timeout        time.Duration `long:"timeout" env:"TIMEOUT" description:"timeout for http requests" default:"5s"`
}

func (g TelegramGroup) build() (notify.Destination, error) {
	lg := cloneLogger(log.Default())
	lg.SetPrefix("[TG] " + lg.Prefix())

	return notify.NewTelegram(notify.TelegramParams{
		ChatID:                g.ChatID,
		Client:                http.Client{Timeout: g.Timeout},
		Token:                 g.Token,
		DisableWebPagePreview: !g.WebPagePreview,
		Log:                   lg,
	}), nil
}

// MattermostHookGroup defines parameters for mattermost hook notifier.
type MattermostHookGroup struct {
	URL     string        `long:"url" env:"URL" description:"url of the mattermost hook"`
	Timeout time.Duration `long:"timeout" env:"TIMEOUT" description:"timeout for http requests" default:"5s"`
}

func (g MattermostHookGroup) build() (notify.Destination, error) {
	lg := cloneLogger(log.Default())
	lg.SetPrefix("[MM_HOOK] " + lg.Prefix())

	return notify.NewMattermost(lg, http.Client{Timeout: g.Timeout}, g.URL), nil
}

// MattermostBotGroup defines parameters for mattermost bot notifier.
type MattermostBotGroup struct {
	BaseURL   string        `long:"base-url" env:"BASE_URL" description:"base url for mattermost API"`
	Token     string        `long:"token" env:"TOKEN" description:"token of the mattermost bot"`
	ChannelID string        `long:"channel-id" env:"CHANNEL_ID" description:"channel id of the mattermost bot"`
	Timeout   time.Duration `long:"timeout" env:"TIMEOUT" description:"timeout for http requests" default:"5s"`
}

func (g MattermostBotGroup) build() (notify.Destination, error) {
	lg := cloneLogger(log.Default())
	lg.SetPrefix("[MM_BOT] " + lg.Prefix())

	return notify.NewMattermostBot(lg, http.Client{Timeout: g.Timeout}, g.BaseURL, g.Token, g.ChannelID)
}

// PostGroup defines parameters for post notifier.
type PostGroup struct {
	URL     string        `long:"url" env:"URL" description:"url to send the release notes"`
	Timeout time.Duration `long:"timeout" env:"TIMEOUT" description:"timeout for http requests" default:"5s"`
}

func (g PostGroup) build() (notify.Destination, error) {
	lg := cloneLogger(log.Default())
	lg.SetPrefix("[POST] " + lg.Prefix())

	return &notify.Post{
		Log:    lg,
		URL:    g.URL,
		Client: &http.Client{Timeout: g.Timeout},
	}, nil
}

// Build builds the notifier.
func (r *NotifyGroup) Build() (destinations notify.Destinations, err error) {
	if r.Stdout {
		destinations = append(destinations, &notify.WriterNotifier{Writer: os.Stdout, Name: "stdout"})
	}
	if r.Stderr {
		destinations = append(destinations, &notify.WriterNotifier{Writer: os.Stderr, Name: "stderr"})
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
		{name: "mattermost-bot", empty: r.MattermostBot.empty(), build: r.MattermostBot.build},
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

func cloneLogger(lg *log.Logger) *log.Logger {
	return log.New(lg.Writer(), lg.Prefix(), lg.Flags())
}

func (g MattermostBotGroup) empty() bool {
	return g.BaseURL == "" || g.Token == "" || g.ChannelID == ""
}
func (g PostGroup) empty() bool           { return g.URL == "" }
func (g MattermostHookGroup) empty() bool { return g.URL == "" }
func (g TelegramGroup) empty() bool       { return g.ChatID == "" || g.Token == "" }
func (g GithubNotifierGroup) empty() bool {
	return g.ReleaseNameTemplate == "" || (g.Repo.FullName == "" && (g.Repo.Owner == "" || g.Repo.Name == ""))
}
