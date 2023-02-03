// Package notify defines interfaces each supported notification destination should implement.
package notify

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
)

//go:generate moq -out mock_destination.go . Destination

// Destination defines interface for a given destination service,
// like telegram, email or stdout.
type Destination interface {
	fmt.Stringer
	// Send the release notes to the destination.
	// tagName is the name of the tag the release notes are for,
	// might be empty and will be ignored by some destinations.
	Send(ctx context.Context, tagName, text string) error
}

// Destinations is an aggregation of notifiers.
type Destinations []Destination

// String returns the names of all destinations.
func (d Destinations) String() string {
	dests := make([]string, len(d))
	for i, dest := range d {
		dests[i] = dest.String()
	}
	return fmt.Sprintf("[%s]", strings.Join(dests, ", "))
}

// Send sends the message to all destinations.
func (d Destinations) Send(ctx context.Context, tagName, text string) error {
	wg := &sync.WaitGroup{}
	wg.Add(len(d))

	errs := make(chan error, len(d))
	for _, dest := range d {
		dest := dest
		go func() {
			defer wg.Done()
			if err := dest.Send(ctx, tagName, text); err != nil {
				errs <- fmt.Errorf("%s: %w", dest, err)
			}
		}()
	}

	wg.Wait()
	close(errs)

	var merr *multierror.Error
	for err := range errs {
		if err != nil {
			merr = multierror.Append(merr, err)
		}
	}

	return merr.ErrorOrNil()
}

func extractBaseURL(url string) string {
	proto := strings.Split(url, "://")
	if len(proto) != 2 {
		return url
	}

	return proto[0] + "://" + strings.Split(proto[1], "/")[0]
}
