package notify

import (
	"context"
	"fmt"
	"io"
)

// WriterNotifier prints the changelog to the specified writer.
// Writer might be os.Stdout, in order to use in pipelines
// and CI/CD.
type WriterNotifier struct {
	Writer io.Writer
	Name   string // used for debugging purposes
}

// String returns the string representation to identify this notifier.
func (w *WriterNotifier) String() string {
	return fmt.Sprintf("writer to %s", w.Name)
}

// Send writes changelog to writer.
func (w *WriterNotifier) Send(_ context.Context, _, text string) error {
	if _, err := io.WriteString(w.Writer, text); err != nil {
		return fmt.Errorf("write release notes: %w", err)
	}
	return nil
}
