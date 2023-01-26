package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/Semior001/releaseit/app/cmd/flg"
)

// ReleaseNotes builds the release-notes from the specified template
// ands sends it to the desired destinations (telegram, stdout (for CI), etc.).
type ReleaseNotes struct {
	Tag     string        `long:"tag" env:"TAG" description:"tag to be released" required:"true"`
	Timeout time.Duration `long:"timeout" env:"TIMEOUT" description:"timeout for assembling the release" default:"5m"`
	flg.ServiceGroup
}

// Execute the release-notes command.
func (r ReleaseNotes) Execute(_ []string) error {
	svc, err := r.ServiceGroup.Build()
	if err != nil {
		return fmt.Errorf("prepare service: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.Timeout)
	defer cancel()

	if err = svc.ReleaseTag(ctx, r.Tag); err != nil {
		return fmt.Errorf("release: %w", err)
	}

	return nil
}
