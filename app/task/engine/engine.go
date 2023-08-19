// Package engine contains implementations of task.Tracker for different task providers.
package engine

import (
	"context"
	"fmt"
	"github.com/Semior001/releaseit/app/task"
	"github.com/samber/lo"
)

//go:generate rm -f mock_interface.go
//go:generate moq -out mock_interface.go . Interface

// Interface defines methods for task tracker engines.
type Interface interface {
	// List lists tasks by their IDs.
	List(ctx context.Context, ids []string) ([]task.Ticket, error)
	// Get returns a single task by its ID.
	Get(ctx context.Context, id string) (task.Ticket, error)
}

// Tracker is a wrapper for task tracker engine with common functions
// for each tracker implementation.
type Tracker struct {
	Engine Interface
}

// List lists tasks by their IDs and parents, if flag is set.
func (s *Tracker) List(ctx context.Context, ids []string, loadParents bool) ([]task.Ticket, error) {
	var result []task.Ticket
	for len(ids) > 0 {
		tickets, err := s.Engine.List(ctx, ids)
		if err != nil {
			return nil, fmt.Errorf("list tickets %s: %w", ids, err)
		}
		result = append(result, tickets...)

		if !loadParents {
			break
		}

		ids = nil
		for _, ticket := range tickets {
			if ticket.ParentID != "" {
				ids = append(ids, ticket.ParentID)
			}
		}
		ids = lo.Uniq(ids)
	}
	return result, nil
}
