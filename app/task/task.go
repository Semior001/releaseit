// Package task contains types and engines to work with tasks.
package task

import (
	"time"
)

// Type specifies the type of the task.
type Type string

const (
	// TypeEpic is an epic task type.
	TypeEpic Type = "epic"
	// TypeTask is a simple task type.
	TypeTask Type = "task"
	// TypeSubtask is a sub-task type.
	TypeSubtask Type = "subtask"
)

// Ticket represents a single task in task tracker.
type Ticket struct {
	ID       string
	ParentID string

	URL          string
	Name         string
	Body         string
	ClosedAt     time.Time
	Author       User
	Assignee     User
	Type         Type
	TypeRaw      string // save raw type in case if user wants to distinguish different raw values
	Flagged      bool
	Watchers     []User
	WatchesCount int
}

// User represents a task tracker user.
type User struct {
	Username string
	Email    string
}
