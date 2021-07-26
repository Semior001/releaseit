package notify

import (
	"context"
	"fmt"
	"io"

	"github.com/Semior001/releaseit/app/store"
	"github.com/Semior001/releaseit/app/store/service"
)

// WriterNotifier prints the changelog to the specified writer.
// Writer might be os.Stdout, in order to use in pipelines
// and CI/CD.
type WriterNotifier struct {
	ReleaseNotesBuilder *service.ReleaseNotesBuilder
	Writer              io.Writer
	Name                string // used for debugging purposes
}

// String returns the string representation to identify this notifier.
func (w *WriterNotifier) String() string {
	return fmt.Sprintf("writer to %s", w.Name)
}

// Send writes changelog to writer.
func (w *WriterNotifier) Send(_ context.Context, changelog store.Changelog) error {
	text, err := w.ReleaseNotesBuilder.Build(changelog)
	if err != nil {
		return fmt.Errorf("build release notes: %w", err)
	}

	if _, err := io.WriteString(w.Writer, text); err != nil {
		return fmt.Errorf("write release notes: %w", err)
	}
	return nil
}
