package flg

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Semior001/releaseit/app/notify"
	"github.com/Semior001/releaseit/app/service/notes"
)

// NotifyGroup defines parameters for the notifier.
type NotifyGroup struct {
	Telegram       TelegramGroup       `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`
	Github         GithubNotifierGroup `group:"github" namespace:"github" env-namespace:"GITHUB"`
	Mattermost     MattermostGroup     `group:"mattermost" namespace:"mattermost" env-namespace:"MATTERMOST"`
	MattermostHook MattermostHookGroup `group:"mattermost-hook" namespace:"mattermost-hook" env-namespace:"MATTERMOST_HOOK"`
	Post           PostGroup           `group:"post" namespace:"post" env-namespace:"POST"`
	Stdout         bool                `long:"stdout" env:"STDOUT" description:"print release notes to stdout"`
	ConfLocation   string              `long:"conf_location" env:"CONF_LOCATION" description:"location to the config file" required:"true"`
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
		{name: "mattermost", empty: r.Mattermost.empty(), build: r.Mattermost.build},
		{name: "mattermost-hook", empty: r.MattermostHook.empty(), build: r.MattermostHook.build},
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

func (g PostGroup) empty() bool           { return g.URL == "" }
func (g MattermostHookGroup) empty() bool { return g.BaseURL == "" || g.ID == "" }
func (g TelegramGroup) empty() bool       { return g.ChatID == "" || g.Token == "" }
func (g MattermostGroup) empty() bool {
	return g.BaseURL == "" || g.ChannelID == "" || g.LoginID == "" || g.Password == ""
}
func (g GithubNotifierGroup) empty() bool {
	return g.ReleaseNameTemplate == "" || g.Repo.Owner == "" || g.Repo.Name == ""
}

// ReleaseNotesBuilder builds the release notes builder.
func (r *NotifyGroup) ReleaseNotesBuilder() (*notes.Builder, error) {
	rnb, err := notes.NewBuilder(r.ConfLocation, r.Extras)
	if err != nil {
		return nil, fmt.Errorf("make release notes builder: %w", err)
	}

	return rnb, nil
}
