package cmd

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/Semior001/releaseit/app/service"
	"github.com/Semior001/releaseit/app/service/eval"
	"github.com/Semior001/releaseit/app/service/notes"
)

// Changelog builds the release-notes from the specified template
// ands sends it to the desired destinations (telegram, stdout (for CI), etc.).
type Changelog struct {
	From           string            `long:"from" env:"FROM" description:"commit ref to start release notes from" default:"{{ previousTag .To (headed (filter semver tags)) }}"`
	To             string            `long:"to" env:"TO" description:"commit ref to end release notes to" default:"{{ last (filter semver tags) }}"`
	Timeout        time.Duration     `long:"timeout" env:"TIMEOUT" description:"timeout for assembling the release" default:"5m"`
	SquashCommitRx string            `long:"squash-commit-rx" env:"SQUASH_COMMIT_RX" description:"regexp to match squash commits" default:"(#\\d+)"`
	ConfLocation   string            `long:"conf-location" env:"CONF_LOCATION" description:"location to the config file" required:"true"`
	Extras         map[string]string `long:"extras" env:"EXTRAS" env-delim:"," description:"extra variables to use in the template"`

	Engine EngineGroup `group:"engine" namespace:"engine" env-namespace:"ENGINE"`
	Notify NotifyGroup `group:"notify" namespace:"notify" env-namespace:"NOTIFY"`
	Task   TaskGroup   `group:"task" namespace:"task" env-namespace:"TASK"`
}

// Execute the release-notes command.
func (r Changelog) Execute(_ []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()

	gitEngine, err := r.Engine.Build(ctx)
	if err != nil {
		return fmt.Errorf("prepare engine: %w", err)
	}

	taskService, err := r.Task.Build(ctx)
	if err != nil {
		return fmt.Errorf("prepare task service: %w", err)
	}

	rnbCfg, err := notes.ConfigFromFile(r.ConfLocation)
	if err != nil {
		return fmt.Errorf("read release notes builder config: %w", err)
	}

	rnbEvaler := &eval.Evaluator{
		Addon: eval.MultiAddon{
			&eval.Git{Engine: gitEngine},
			&eval.Task{Tracker: taskService},
			&notes.EvalAddon{TaskTracker: taskService},
		},
	}

	if err = rnbEvaler.Validate(rnbCfg.Template); err != nil {
		return fmt.Errorf("release notes template is invalid: %w", err)
	}

	rnb, err := notes.NewBuilder(rnbCfg, rnbEvaler, r.Extras)
	if err != nil {
		return fmt.Errorf("prepare release notes builder: %w", err)
	}

	notif, err := r.Notify.Build()
	if err != nil {
		return fmt.Errorf("prepare notifier: %w", err)
	}

	rx, err := regexp.Compile(r.SquashCommitRx)
	if err != nil {
		return fmt.Errorf("compile squash commit regexp: %w", err)
	}

	svc := &service.Service{
		Evaluator:             &eval.Evaluator{Addon: &eval.Git{Engine: gitEngine}},
		Engine:                gitEngine,
		ReleaseNotesBuilder:   rnb,
		Notifier:              notif,
		SquashCommitMessageRx: rx,
	}

	if err = svc.Changelog(ctx, r.From, r.To); err != nil {
		return fmt.Errorf("build changelog: %w", err)
	}

	return nil
}
