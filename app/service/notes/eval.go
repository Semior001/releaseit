package notes

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/Semior001/releaseit/app/git"
	"github.com/Semior001/releaseit/app/task"
	tengine "github.com/Semior001/releaseit/app/task/engine"
	"github.com/samber/lo"
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
		"listTaskUsers":    e.listTaskUsers,
		"listPRs":          e.listPRs,
		"mdTaskLink":       e.mdTaskLink,
		"brackets":         e.brackets,
		"log":              e.log,
	}, nil
}

// listTaskUsers returns a string with task users.
// KVargs:
// 0: prefix for users
// 1: key for author
// 2: key for assignee
// 3: key for watchers
func (e *EvalAddon) listTaskUsers(obj any, args ...string) (string, error) {
	tg, ok := obj.(interface{ GetTicket() task.Ticket })
	if !ok {
		return "", fmt.Errorf("expected task.Ticket, got %T", obj)
	}

	t := tg.GetTicket()

	defaultKeys := []string{"", "author", "assignee", "watchers"}

	for i := range args {
		if i < len(defaultKeys) {
			defaultKeys[i] = args[i]
		}
	}

	uPrefix, authorKey, assigneeKey, watcherKey := defaultKeys[0], defaultKeys[1], defaultKeys[2], defaultKeys[3]
	var parts []string
	if t.Author.Username != "" {
		parts = append(parts, fmt.Sprintf("%s %s%s", authorKey, uPrefix, t.Author.Username))
	}
	if t.Assignee.Username != "" {
		parts = append(parts, fmt.Sprintf("%s %s%s", assigneeKey, uPrefix, t.Assignee.Username))
	}
	if len(t.Watchers) > 0 {
		var usernames []string
		for _, user := range t.Watchers {
			usernames = append(usernames, fmt.Sprintf("%s%s", uPrefix, user.Username))
		}
		parts = append(parts, fmt.Sprintf("%s: %s", watcherKey, strings.Join(usernames, ", ")))
	}

	return strings.Join(parts, ", "), nil
}

func (e *EvalAddon) log(msg string, args ...any) string {
	log.Printf("[DEBUG][evaluator] "+msg, args...)
	return ""
}

func (e *EvalAddon) brackets(s string, square ...bool) string {
	if s == "" {
		return ""
	}
	if len(square) > 0 && square[0] {
		return fmt.Sprintf("[%s]", s)
	}
	return fmt.Sprintf("(%s)", s)
}

func (e *EvalAddon) listPRs(prs []git.PullRequest) string {
	var parts []string
	for _, pr := range prs {
		parts = append(parts, fmt.Sprintf("[%s](%s)", pr.Title, pr.URL))
	}
	return strings.Join(parts, ", ")
}

func (e *EvalAddon) mdTaskLink(obj any) (string, error) {
	tg, ok := obj.(interface{ GetTicket() task.Ticket })
	if !ok {
		return "", fmt.Errorf("expected task.Ticket, got %T", obj)
	}

	t := tg.GetTicket()

	return fmt.Sprintf("[%s](%s)", t.ID, t.URL), nil
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

func (e *EvalAddon) loadTicketsTree(ctx context.Context) func(string, bool, []git.PullRequest, []git.Commit) (LoadedTree, error) {
	return func(ticketIDRx string, loadParents bool, prs []git.PullRequest, commits []git.Commit) (LoadedTree, error) {
		rx, err := regexp.Compile(ticketIDRx)
		if err != nil {
			return LoadedTree{}, fmt.Errorf("compile regexp: %w", err)
		}

		ticketPRs := map[string][]git.PullRequest{} // ticketID -> PR index
		ticketCommits := map[string][]git.Commit{}  // ticketID -> commit hash
		var unattachedPRs []git.PullRequest
		var unattachedCommits []git.Commit

		for _, pr := range prs {
			submatches := rx.FindAllStringSubmatch(pr.Title, -1)
			if len(submatches) == 0 {
				unattachedPRs = append(unattachedPRs, pr)
				continue
			}

			for _, submatch := range submatches {
				ticketID := submatch[1]
				ticketPRs[ticketID] = append(ticketPRs[ticketID], pr)
			}
		}

		for _, commit := range commits {
			submatches := rx.FindAllStringSubmatch(commit.Message, -1)
			if len(submatches) == 0 {
				unattachedCommits = append(unattachedCommits, commit)
				continue
			}

			for _, submatch := range submatches {
				ticketID := submatch[1]
				ticketCommits[ticketID] = append(ticketCommits[ticketID], commit)
			}
		}

		tickets, err := e.TaskTracker.List(ctx,
			append(lo.Keys(ticketPRs), lo.Keys(ticketCommits)...),
			loadParents,
		)
		if err != nil {
			return LoadedTree{}, fmt.Errorf("load tickets: %w", err)
		}

		tree, err := e.buildTicketsTree(tickets)
		if err != nil {
			return LoadedTree{}, fmt.Errorf("build tickets tree: %w", err)
		}
		addPRs(ticketPRs, tree)
		addCommits(ticketCommits, tree)

		return LoadedTree{Roots: tree, UnattachedPRs: unattachedPRs, UnattachedCommits: unattachedCommits}, nil
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

func addCommits(commits map[string][]git.Commit, tree []*TicketNode) {
	for _, node := range tree {
		if c, ok := commits[node.ID]; ok {
			node.Commits = c
		}
		addCommits(commits, node.Children)
	}
}

// LoadedTree is a tree of tickets with their children and PRs.
type LoadedTree struct {
	Roots             []*TicketNode
	UnattachedPRs     []git.PullRequest
	UnattachedCommits []git.Commit
}

// TicketNode is a representation of a ticket with its children.
type TicketNode struct {
	task.Ticket
	Children []*TicketNode
	PRs      []git.PullRequest
	Commits  []git.Commit
}

// GetTicket returns the ticket.
// FIXME: this is ugly, but it's needed to match the interface for embedded structs.
func (t *TicketNode) GetTicket() task.Ticket { return t.Ticket }
