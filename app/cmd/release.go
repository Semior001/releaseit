package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Semior001/releaseit/app/notify"
	"github.com/Semior001/releaseit/app/store/engine"
	"github.com/Semior001/releaseit/app/store/service"
	"gopkg.in/yaml.v3"
)

// ReleaseNotes builds the release-notes from the specified template
// ands sends it to the desired destinations (telegram, stdout (for CI), etc.).
type ReleaseNotes struct {
	ConfLocation string      `long:"conf_location" env:"CONF_LOCATION" description:"location to the config file" required:"true"`
	Github       GithubGroup `group:"github" namespace:"github" env-namespace:"GITHUB"`
}

// Execute the release-notes command.
func (r ReleaseNotes) Execute(_ []string) error {
	cfg, err := parseCfg(r.ConfLocation)
	if err != nil {
		return fmt.Errorf("parse release-notes builder config: %w", err)
	}

	releaseNotesBuilder, err := service.NewChangelogBuilder(cfgToServiceParams(cfg))
	if err != nil {
		return fmt.Errorf("initialize changelog builder: %w", err)
	}

	httpCl := http.Client{Timeout: 5 * time.Second}

	notifySrv := &notify.Service{
		Log: log.Default(),
		Destinations: []notify.Destination{
			&notify.WriterNotifier{
				ReleaseNotesBuilder: releaseNotesBuilder,
				Writer:              os.Stdout,
				Name:                "stdout",
			},
		},
	}

	srv := service.Service{
		Engine: engine.NewGithub(r.Github.Repo.Owner, r.Github.Repo.Name, httpCl, engine.BasicAuth{
			Username: r.Github.BasicAuth.Username,
			Password: r.Github.BasicAuth.Password,
		}),
		Notifier: notifySrv,
	}

	if err = srv.Release(context.Background()); err != nil {
		return fmt.Errorf("release: %w", err)
	}

	return nil
}

type cfg struct {
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

func (c cfg) Validate() error {
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

func parseCfg(cfgPath string) (cfg, error) {
	bytes, err := os.ReadFile(cfgPath)
	if err != nil {
		return cfg{}, fmt.Errorf("open file: %w", err)
	}

	var res cfg

	if err = yaml.Unmarshal(bytes, &res); err != nil {
		return cfg{}, fmt.Errorf("parse yaml: %w", err)
	}

	// validating config
	if err = res.Validate(); err != nil {
		return cfg{}, fmt.Errorf("config is invalid: %w", err)
	}

	return res, nil
}

func cfgToServiceParams(cfg cfg) service.Params {
	res := service.Params{
		Template:     cfg.Template,
		IgnoreLabels: cfg.IgnoreLabels,
		Categories:   make([]service.Category, len(cfg.Categories)),
		UnusedTitle:  cfg.UnusedTitle,
		SortField:    cfg.SortField,
	}

	for i, category := range cfg.Categories {
		res.Categories[i] = service.Category{
			Title:  category.Title,
			Labels: category.Labels,
		}
	}

	return res
}
