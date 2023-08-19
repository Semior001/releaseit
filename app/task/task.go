// Package task contains types and engines to work with tasks.
package task

import (
	"time"
)

// Ticket represents a single task in task tracker.
type Ticket struct {
	ID       string
	ParentID string

	Name     string
	Body     string
	ClosedAt time.Time
	Author   User
	Assignee User
}

// User represents a task tracker user.
type User struct {
	Username string
	Email    string
}
