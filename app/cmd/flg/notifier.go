package flg

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/Semior001/releaseit/app/config"
	"github.com/Semior001/releaseit/app/notify"
	"github.com/Semior001/releaseit/app/service"
	"github.com/samber/lo"
)

// NotifyGroup defines parameters for the notifier.
type NotifyGroup struct {
	Telegram       TelegramGroup       `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`
	Github         GithubNotifierGroup `group:"github" namespace:"github" env-namespace:"GITHUB"`
	Mattermost     MattermostGroup     `group:"mattermost" namespace:"mattermost" env-namespace:"MATTERMOST"`
	MattermostHook MattermostHookGroup `group:"mattermost-hook" namespace:"mattermost-hook" env-namespace:"MATTERMOST_HOOK"`
	Stdout         bool                `long:"stdout" env:"STDOUT" description:"print release notes to stdout"`
	ConfLocation   string              `long:"conf_location" env:"CONF_LOCATION" description:"location to the config file"`
	Extras         map[string]string   `long:"extras" env:"EXTRAS" env-delim:"," description:"extra variables to use in the template"`
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

// Empty returns true if the argument group is empty.
func (g GithubNotifierGroup) Empty() bool {
	return g.ReleaseNameTemplate == "" || g.GithubGroup.Empty()
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

// Empty returns true if the config group is not filled.
func (g TelegramGroup) Empty() bool {
	return g.ChatID == "" || g.Token == ""
}

// MattermostGroup defines parameters for mattermost notifier.
type MattermostGroup struct {
	BaseURL   string        `long:"base_url" env:"BASE_URL" description:"base url of the mattermost server"`
	ChannelID string        `long:"channel_id" env:"CHANNEL_ID" description:"id of the channel, where the release notes will be sent"`
	LoginID   string        `long:"login_id" env:"LOGIN_ID" description:"login id of the user, who will send the release notes"`
	Password  string        `long:"password" env:"PASSWORD" description:"password of the user, who will send the release notes"`
	LDAP      bool          `long:"ldap" env:"LDAP" description:"use ldap auth"`
	Timeout   time.Duration `long:"timeout" env:"TIMEOUT" description:"timeout for http requests" default:"5s"`
}

func (g MattermostGroup) build() (notify.Destination, error) {
	return notify.NewMattermostBot(notify.MattermostBotParams{
		Client:    http.Client{Timeout: g.Timeout},
		BaseURL:   g.BaseURL,
		ChannelID: g.ChannelID,
		LoginID:   g.LoginID,
		Password:  g.Password,
		LDAP:      g.LDAP,
	})
}

// Empty returns true if the config group is not filled.
func (g MattermostGroup) Empty() bool {
	return g.BaseURL == "" || g.ChannelID == "" || g.LoginID == "" || g.Password == ""
}

// MattermostHookGroup defines parameters for mattermost hook notifier.
type MattermostHookGroup struct {
	BaseURL string        `long:"base_url" env:"BASE_URL" description:"base url of the mattermost server"`
	ID      string        `long:"id" env:"ID" description:"id of the hook, where the release notes will be sent"`
	Timeout time.Duration `long:"timeout" env:"TIMEOUT" description:"timeout for http requests" default:"5s"`
}

func (g MattermostHookGroup) build() (notify.Destination, error) {
	return notify.NewMattermostHook(
		http.Client{Timeout: g.Timeout},
		g.BaseURL,
		g.ID,
	), nil
}

// Empty returns true if the config group is not filled.
func (g MattermostHookGroup) Empty() bool {
	return g.BaseURL == "" || g.ID == ""
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
		{name: "telegram", empty: r.Telegram.Empty(), build: r.Telegram.build},
		{name: "github", empty: r.Github.Empty(), build: r.Github.build},
		{name: "mattermost", empty: r.Mattermost.Empty(), build: r.Mattermost.build},
		{name: "mattermost-hook", empty: r.MattermostHook.Empty(), build: r.MattermostHook.build},
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

// ReleaseNotesBuilder builds the release notes builder.
func (r *NotifyGroup) ReleaseNotesBuilder() (*service.ReleaseNotesBuilder, error) {
	cfg, err := config.Read(r.ConfLocation)
	if err != nil {
		return nil, fmt.Errorf("parse release-notes builder config: %w", err)
	}

	rnb := &service.ReleaseNotesBuilder{
		Template:     cfg.Template,
		IgnoreLabels: cfg.IgnoreLabels,
		Categories:   make([]service.Category, len(cfg.Categories)),
		UnusedTitle:  cfg.UnusedTitle,
		SortField:    cfg.SortField,
		Extras:       r.Extras,
	}

	for i, category := range cfg.Categories {
		rnb.Categories[i] = service.Category{
			Title:        category.Title,
			Labels:       category.Labels,
			BranchRegexp: lo.Ternary(category.Branch == "", nil, regexp.MustCompile(category.Branch)),
		}
	}

	return rnb, nil
}
