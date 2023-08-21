package eval

import (
	"context"
	"github.com/Semior001/releaseit/app/task"
	"github.com/Semior001/releaseit/app/task/engine"
	"text/template"
)

// Task is an addon to the Eval service that allows to
// list and manage tasks from task tracker.
type Task struct {
	Tracker *engine.Tracker
}

// String returns the name of the template addon.
func (t *Task) String() string { return "task" }

// Funcs returns a map of functions for use in templates.
func (t *Task) Funcs(ctx context.Context) (template.FuncMap, error) {
	return template.FuncMap{
		"getTicket":   t.getTicket(ctx),
		"listTickets": t.listTickets(ctx),
	}, nil
}

func (t *Task) getTicket(ctx context.Context) func(id string) (task.Ticket, error) {
	return func(id string) (task.Ticket, error) {
		return t.Tracker.Get(ctx, id)
	}
}

func (t *Task) listTickets(ctx context.Context) func(ids []string, loadParents bool) ([]task.Ticket, error) {
	return func(ids []string, loadParents bool) ([]task.Ticket, error) {
		return t.Tracker.List(ctx, ids, loadParents)
	}
}
