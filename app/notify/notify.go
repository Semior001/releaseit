package notify

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Semior001/releaseit/app/git"
	"golang.org/x/sync/errgroup"
)

// Destination defines interface for a given destination service,
// like telegram, email or stdout.
type Destination interface {
	fmt.Stringer
	Send(ctx context.Context, changelog git.Changelog) error
}

// Service delivers changelog to multiple destinations.
type Service struct {
	l    *log.Logger
	dest []Destination
}

// Params describes parameters to initialize notifier service.
type Params struct {
	Log          *log.Logger
	Destinations []Destination
}

// NewService makes new instance of Service.
func NewService(params Params) *Service {
	svc := &Service{
		l:    params.Log,
		dest: params.Destinations,
	}

	svc.l.Printf("[INFO] initialized notifier service: %s", svc.String())

	return svc
}

// String used for debugging purposes.
func (s *Service) String() string {
	dests := make([]string, len(s.dest))
	for i, dest := range s.dest {
		dests[i] = dest.String()
	}
	return fmt.Sprintf("aggregated notifier with next destinations: [%s]", strings.Join(dests, ", "))
}

// Send sends the changelog to all destinations.
func (s *Service) Send(ctx context.Context, changelog git.Changelog) error {
	eg, nestedCtx := errgroup.WithContext(ctx)
	for _, destination := range s.dest {
		destination := destination
		eg.Go(func() error {
			if err := destination.Send(nestedCtx, changelog); err != nil {
				s.l.Printf("[WARN] failed to send changelog to destination %s: %v", destination.String(), err)
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
