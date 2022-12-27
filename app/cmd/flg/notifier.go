package flg

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Semior001/releaseit/app/config"
	"github.com/Semior001/releaseit/app/notify"
	"github.com/Semior001/releaseit/app/service"
)

// NotifyGroup defines parameters for the notifier.
type NotifyGroup struct {
	Telegram     TelegramGroup       `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`
	Github       GithubNotifierGroup `group:"github" namespace:"github" env-namespace:"GITHUB"`
	Stdout       bool                `long:"stdout" env:"STDOUT" description:"print release notes to stdout"`
	ConfLocation string              `long:"conf_location" env:"CONF_LOCATION" description:"location to the config file"`
}

// GithubNotifierGroup defines parameters to make release in the github.
type GithubNotifierGroup struct {
	GithubGroup
	ReleaseNameTemplate string `long:"release_name_tmpl" env:"RELEASE_NAME_TMPL" description:"template for release name"`
}

// Empty returns true if the argument group is empty.
func (g GithubNotifierGroup) Empty() bool {
	return g.ReleaseNameTemplate == "" || g.GithubGroup.Empty()
}

// TelegramGroup defines parameters for telegram notifier.
type TelegramGroup struct {
	ChatID         string `long:"chat_id" env:"CHAT_ID" description:"id of the chat, where the release notes will be sent"`
	Token          string `long:"token" env:"TOKEN" description:"bot token"`
	WebPagePreview bool   `long:"web_page_preview" env:"WEB_PAGE_PREVIEW" description:"request telegram to preview for web links"`
}

// Empty returns true if the config group is not filled.
func (g TelegramGroup) Empty() bool {
	return g.ChatID == "" || g.Token == ""
}

// Build builds the notifier.
func (r *NotifyGroup) Build() (destinations notify.Destinations, err error) {
	logger := log.Default()

	if r.Stdout {
		destinations = append(destinations, &notify.WriterNotifier{Writer: os.Stdout, Name: "stdout"})
	}

	if !r.Telegram.Empty() {
		destinations = append(destinations, notify.NewTelegram(notify.TelegramParams{
			Log:                   logger,
			ChatID:                r.Telegram.ChatID,
			Client:                http.Client{Timeout: 5 * time.Second},
			Token:                 r.Telegram.Token,
			DisableWebPagePreview: !r.Telegram.WebPagePreview,
		}))
	}

	if !r.Github.Empty() {
		gh, err := notify.NewGithub(notify.GithubParams{
			Owner:               r.Github.Repo.Owner,
			Name:                r.Github.Repo.Name,
			BasicAuthUsername:   r.Github.BasicAuth.Username,
			BasicAuthPassword:   r.Github.BasicAuth.Password,
			HTTPClient:          http.Client{Timeout: 5 * time.Second},
			ReleaseNameTmplText: r.Github.ReleaseNameTemplate,
		})
		if err != nil {
			return nil, fmt.Errorf("make github relases notifier: %w", err)
		}

		destinations = append(destinations, gh)
	}

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
	}

	for i, category := range cfg.Categories {
		rnb.Categories[i] = service.Category{
			Title:  category.Title,
			Labels: category.Labels,
		}
	}

	return rnb, nil
}
