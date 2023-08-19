package notes

import (
	"context"
	"fmt"
	"github.com/Semior001/releaseit/app/task"
	"sort"
	"text/template"
)

// EvalAddon is an addon to evaluator, to be used in release notes template.
type EvalAddon struct{}

// String returns addon name.
func (e *EvalAddon) String() string { return "release-notes" }

// Funcs returns a map of functions to be used in release notes template.
func (e *EvalAddon) Funcs(_ context.Context) (template.FuncMap, error) {
	return template.FuncMap{"treeTickets": e.treeTickets}, nil
}

func (e *EvalAddon) treeTickets(tickets []task.Ticket) (roots []*TicketNode, err error) {
	ticketMap := make(map[string]*TicketNode)
	for _, ticket := range tickets {
		ticketMap[ticket.ID] = &TicketNode{Ticket: ticket}
	}

	for _, ticket := range ticketMap {
		if ticket.ParentID == "" {
			roots = append(roots, ticket)
			continue
		}

		parent, ok := ticketMap[ticket.ParentID]
		if !ok {
			return nil, fmt.Errorf("ticket %s has unknown parent %s", ticket.ID, ticket.ParentID)
		}

		parent.Children = append(parent.Children, ticket)
	}

	var sortTickets func(nodes []*TicketNode)
	sortTickets = func(nodes []*TicketNode) {
		sort.Slice(nodes, func(i, j int) bool {
			return nodes[i].ID < nodes[j].ID
		})
		for _, node := range nodes {
			sortTickets(node.Children)
		}
	}
	sortTickets(roots)

	return roots, nil
}

// TicketNode is a representation of a ticket with its children.
type TicketNode struct {
	task.Ticket
	Children []*TicketNode
}
