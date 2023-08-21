package notes

import (
	"context"
	"fmt"
	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/task"
	tengine "github.com/Semior001/releaseit/app/task/engine"
	"github.com/samber/lo"
	"regexp"
	"sort"
	"text/template"
)

// EvalAddon is an addon to evaluator, to be used in release notes template.
type EvalAddon struct {
	TaskTracker *tengine.Tracker
}

// String returns addon name.
func (e *EvalAddon) String() string { return "release-notes" }

// Funcs returns a map of functions to be used in release notes template.
func (e *EvalAddon) Funcs(ctx context.Context) (template.FuncMap, error) {
	return template.FuncMap{
		"buildTicketsTree": e.buildTicketsTree,
		"loadTicketsTree":  e.loadTicketsTree(ctx),
	}, nil
}

func (e *EvalAddon) buildTicketsTree(tickets []task.Ticket) (roots []*TicketNode, err error) {
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

	sortTicketNodes(roots)

	return roots, nil
}

func (e *EvalAddon) loadTicketsTree(ctx context.Context) func(string, bool, []git.PullRequest) (LoadedTree, error) {
	return func(ticketIDRx string, loadParents bool, prs []git.PullRequest) (LoadedTree, error) {
		rx, err := regexp.Compile(ticketIDRx)
		if err != nil {
			return LoadedTree{}, fmt.Errorf("compile regexp: %w", err)
		}

		ticketPRs := map[string][]git.PullRequest{} // ticketID -> PR index
		var unattached []git.PullRequest
		for _, pr := range prs {
			submatches := rx.FindAllStringSubmatch(pr.Title, -1)
			if len(submatches) == 0 {
				unattached = append(unattached, pr)
				continue
			}

			for _, submatch := range submatches {
				ticketID := submatch[1]
				ticketPRs[ticketID] = append(ticketPRs[ticketID], pr)
			}
		}

		tickets, err := e.TaskTracker.List(ctx, lo.Keys(ticketPRs), loadParents)
		if err != nil {
			return LoadedTree{}, fmt.Errorf("load tickets: %w", err)
		}

		tree, err := e.buildTicketsTree(tickets)
		if err != nil {
			return LoadedTree{}, fmt.Errorf("build tickets tree: %w", err)
		}
		addPRs(ticketPRs, tree)

		return LoadedTree{Roots: tree, Unattached: unattached}, nil
	}
}

func sortTicketNodes(nodes []*TicketNode) {
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	for _, node := range nodes {
		sortTicketNodes(node.Children)
	}
}

func addPRs(ticketPRs map[string][]git.PullRequest, nodes []*TicketNode) {
	for _, node := range nodes {
		if prs, ok := ticketPRs[node.ID]; ok {
			node.PRs = append(node.PRs, prs...)
		}
		addPRs(ticketPRs, node.Children)
	}
}

// LoadedTree is a tree of tickets with their children and PRs.
type LoadedTree struct {
	Roots      []*TicketNode
	Unattached []git.PullRequest
}

// TicketNode is a representation of a ticket with its children.
type TicketNode struct {
	task.Ticket
	Children []*TicketNode
	PRs      []git.PullRequest
}
