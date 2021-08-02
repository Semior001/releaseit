package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Semior001/releaseit/app/notify"
	"github.com/Semior001/releaseit/app/store/engine"
	"github.com/Semior001/releaseit/app/store/service"
)

// ReleaseNotes builds the release-notes from the specified template
// ands sends it to the desired destinations (telegram, stdout (for CI), etc.).
type ReleaseNotes struct {
	Tag    string `long:"tag" env:"TAG" description:"tag to be released" required:"true"`
	Engine struct {
		Type   string      `long:"type" env:"TYPE" choice:"github" description:"type of the repository engine" required:"true"`
		Github GithubGroup `group:"github" namespace:"github" env-namespace:"GITHUB"`
	} `group:"engine" namespace:"engine" env-namespace:"ENGINE"`
	Notify struct {
		Telegram TelegramGroup       `group:"telegram" namespace:"telegram" env-namespace:"TELEGRAM"`
		Github   GithubNotifierGroup `group:"github" namespace:"github" env-namespace:"GITHUB"`
		Stdout   StdoutGroup         `group:"stdout" namespace:"stdout" env-namespace:"STDOUT"`
	} `group:"notify" namespace:"notify" env-namespace:"NOTIFY"`
}

// Execute the release-notes command.
func (r ReleaseNotes) Execute(_ []string) error {
	eng, err := r.makeEngine()
	if err != nil {
		return err
	}

	notif, err := r.makeNotifier()
	if err != nil {
		return err
	}

	if err = service.NewService(eng, notif).Release(context.Background(), r.Tag); err != nil {
		return fmt.Errorf("release: %w", err)
	}

	return nil
}

func (r ReleaseNotes) makeEngine() (engine.Interface, error) {
	switch r.Engine.Type {
	case "github":
		return engine.NewGithub(
			r.Engine.Github.Repo.Owner,
			r.Engine.Github.Repo.Name,
			r.Engine.Github.BasicAuth.Username,
			r.Engine.Github.BasicAuth.Password,
			http.Client{Timeout: 5 * time.Second},
		)
	}
	return nil, fmt.Errorf("unsupported repository engine type %s", r.Engine.Type)
}

func (r ReleaseNotes) makeNotifier() (*notify.Service, error) {
	logger := log.Default()

	var destinations []notify.Destination

	if !r.Notify.Stdout.Empty() {
		changelogBuilder, err := makeChangelogBuilder(r.Notify.Stdout.ConfLocation)
		if err != nil {
			return nil, fmt.Errorf("make changelog builder for stdout: %w", err)
		}

		destinations = append(destinations, &notify.WriterNotifier{
			ReleaseNotesBuilder: changelogBuilder,
			Writer:              os.Stdout,
			Name:                "stdout",
		})
	}

	if !r.Notify.Telegram.Empty() {
		changelogBuilder, err := makeChangelogBuilder(r.Notify.Telegram.ConfLocation)
		if err != nil {
			return nil, fmt.Errorf("make changelog builder for telegram: %w", err)
		}

		destinations = append(destinations, notify.NewTelegram(notify.TelegramParams{
			ReleaseNotesBuilder:   changelogBuilder,
			Log:                   logger,
			ChatID:                r.Notify.Telegram.ChatID,
			Client:                http.Client{Timeout: 5 * time.Second},
			Token:                 r.Notify.Telegram.Token,
			DisableWebPagePreview: !r.Notify.Telegram.WebPagePreview,
		}))
	}

	if !r.Notify.Github.Empty() {
		changelogBuilder, err := makeChangelogBuilder(r.Notify.Github.ConfLocation)
		if err != nil {
			return nil, fmt.Errorf("make changelog builder for github releases: %w", err)
		}

		gh, err := notify.NewGithub(notify.GithubParams{
			Owner:               r.Notify.Github.Repo.Owner,
			Name:                r.Notify.Github.Repo.Name,
			BasicAuthUsername:   r.Notify.Github.BasicAuth.Username,
			BasicAuthPassword:   r.Notify.Github.BasicAuth.Password,
			HTTPClient:          http.Client{Timeout: 5 * time.Second},
			ReleaseNotesBuilder: changelogBuilder,
			ReleaseNameTmplText: r.Notify.Github.ReleaseNameTemplate,
		})
		if err != nil {
			return nil, fmt.Errorf("make github relases notifier: %w", err)
		}

		destinations = append(destinations, gh)
	}

	return notify.NewService(notify.Params{Log: logger, Destinations: destinations}), nil
}

func makeChangelogBuilder(cfgLocation string) (*service.ReleaseNotesBuilder, error) {
	cfg, err := parseCfg(cfgLocation)
	if err != nil {
		return nil, fmt.Errorf("parse release-notes builder config: %w", err)
	}

	params := service.Params{
		Template:     cfg.Template,
		IgnoreLabels: cfg.IgnoreLabels,
		Categories:   make([]service.Category, len(cfg.Categories)),
		UnusedTitle:  cfg.UnusedTitle,
		SortField:    cfg.SortField,
	}

	for i, category := range cfg.Categories {
		params.Categories[i] = service.Category{
			Title:  category.Title,
			Labels: category.Labels,
		}
	}

	releaseNotesBuilder, err := service.NewChangelogBuilder(params)
	if err != nil {
		return nil, fmt.Errorf("initialize changelog builder: %w", err)
	}

	return releaseNotesBuilder, nil
}
