// Package task contains types and engines to work with tasks.
package task

import (
	"context"
	"time"
)

// Tracker defines methods for task tracker engines.
type Tracker interface {
	// List lists tasks by their IDs.
	List(ctx context.Context, ids []string) ([]Ticket, error)
	// Get returns a single task by its ID.
	Get(ctx context.Context, id string) (Ticket, error)
}

// Ticket represents a single task in task tracker.
type Ticket struct {
	ID       string
	Name     string
	Body     string
	ClosedAt time.Time
	Parent   *Ticket
	Author   User
	Assignee User

	Children []Ticket // filled manually, if requested, used in templates only
}

// User represents a task tracker user.
type User struct {
	Username string
	Email    string
}
