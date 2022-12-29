package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/Semior001/releaseit/app/cmd/flg"
	"github.com/Semior001/releaseit/app/service"
)

// ReleaseNotes builds the release-notes from the specified template
// ands sends it to the desired destinations (telegram, stdout (for CI), etc.).
type ReleaseNotes struct {
	Tag     string          `long:"tag" env:"TAG" description:"tag to be released" required:"true"`
	Timeout time.Duration   `long:"timeout" env:"TIMEOUT" description:"timeout for assembling the release" default:"5m"`
	Engine  flg.EngineGroup `group:"engine" namespace:"engine" env-namespace:"ENGINE"`
	Notify  flg.NotifyGroup `group:"notify" namespace:"notify" env-namespace:"NOTIFY"`
}

// Execute the release-notes command.
func (r ReleaseNotes) Execute(_ []string) error {
	eng, err := r.Engine.Build()
	if err != nil {
		return err
	}

	notif, err := r.Notify.Build()
	if err != nil {
		return err
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

	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()

	if err = svc.ReleaseTag(ctx, r.Tag); err != nil {
		return fmt.Errorf("release: %w", err)
	}

	return nil
}
