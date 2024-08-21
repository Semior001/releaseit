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
	ID       string `yaml:"id"`
	ParentID string `yaml:"parent_id"`

	URL          string    `yaml:"url"`
	Name         string    `yaml:"name"`
	Body         string    `yaml:"body"`
	ClosedAt     time.Time `yaml:"closed_at"`
	Author       User      `yaml:"author"`
	Assignee     User      `yaml:"assignee"`
	Type         Type      `yaml:"type"`
	TypeRaw      string    `yaml:"type_raw"` // save raw type in case if user wants to distinguish different raw values
	Flagged      bool      `yaml:"flagged"`
	Watchers     []User    `yaml:"watchers"`
	WatchesCount int       `yaml:"watches_count"`
}

// GetTicket returns the ticket itself.
// FIXME: this is ugly, but it's needed to match the interface for embedded structs.
func (t Ticket) GetTicket() Ticket { return t }

// User represents a task tracker user.
type User struct {
	Username string
	Email    string
}
