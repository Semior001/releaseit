package notify

import (
	"context"
	"fmt"
	"log"

	"github.com/Semior001/releaseit/app/store"
	"golang.org/x/sync/errgroup"
)

// Destination defines interface for a given destination service,
// like telegram, email or stdout.
type Destination interface {
	fmt.Stringer
	Send(ctx context.Context, changelog store.Changelog) error
}

// Service delivers changelog to multiple destinations.
type Service struct {
	Log          *log.Logger
	Destinations []Destination
}

// String used for debugging purposes.
func (s *Service) String() string {
	return fmt.Sprintf("aggregated notifier with next notifiers: %s", s.Destinations)
}

// Send sends the changelog to all destinations.
func (s *Service) Send(ctx context.Context, changelog store.Changelog) error {
	eg, nestedCtx := errgroup.WithContext(ctx)
	for _, destination := range s.Destinations {
		destination := destination
		eg.Go(func() error {
			if err := destination.Send(nestedCtx, changelog); err != nil {
				s.Log.Printf("[WARN] failed to send changelog to destination %s: %v", destination.String(), err)
				return err
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("notify: %w", err)
	}
	return nil
}
