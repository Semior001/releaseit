package cmd

import (
	"context"
	"fmt"

	"github.com/Semior001/releaseit/app/cmd/flg"
	"github.com/Semior001/releaseit/app/service"
)

// Changelog builds the release-notes from the specified template
// ands sends it to the desired destinations (telegram, stdout (for CI), etc.).
type Changelog struct {
	From   string          `long:"from" env:"FROM" description:"sha to start release notes from" required:"true"`
	To     string          `long:"to" env:"TO" description:"sha to end release notes to" required:"true"`
	Engine flg.EngineGroup `group:"engine" namespace:"engine" env-namespace:"ENGINE"`
	Notify flg.NotifyGroup `group:"notify" namespace:"notify" env-namespace:"NOTIFY"`
}

// Execute the release-notes command.
func (r Changelog) Execute(_ []string) error {
	eng, err := r.Engine.Build()
	if err != nil {
		return fmt.Errorf("prepare engine: %w", err)
	}

	notif, err := r.Notify.Build()
	if err != nil {
		return fmt.Errorf("prepare notifier: %w", err)
	}

	rnb, err := r.Notify.ReleaseNotesBuilder()
	if err != nil {
		return fmt.Errorf("prepare release notes builder: %w", err)
	}

	svc := &service.Service{
		Engine:              eng,
		ReleaseNotesBuilder: rnb,
		Notifier:            notif,
	}

	if err = svc.ReleaseBetween(context.Background(), r.From, r.To); err != nil {
		return fmt.Errorf("assemble changelog: %w", err)
	}

	return nil
}
