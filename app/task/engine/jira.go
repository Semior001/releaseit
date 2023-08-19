package engine

import (
	"context"
	"fmt"
	"github.com/Semior001/releaseit/app/task"
	"github.com/andygrunwald/go-jira"
	"github.com/go-pkgz/requester"
	"github.com/go-pkgz/requester/middleware"
	"net/http"
	"strings"
	"time"
)

// Jira is a Jira task tracker engine.
type Jira struct {
	cl *jira.Client
}

// JiraParams is a set of parameters for Jira engine.
type JiraParams struct {
	URL        string
	Token      string
	HTTPClient http.Client
}

// NewJira creates a new Jira engine.
func NewJira(params JiraParams) (*Jira, error) {
	rq := requester.New(params.HTTPClient,
		middleware.Header("Authorization", "Bearer "+params.Token),
	)

	cl, err := jira.NewClient(rq, params.URL)
	if err != nil {
		return nil, err
	}

	return &Jira{cl: cl}, nil
}

// List lists tasks from the provided project by their IDs.
func (j *Jira) List(ctx context.Context, keys []string) ([]task.Ticket, error) {
	query := fmt.Sprintf("key in (%s)", strings.Join(keys, ","))
	issues, _, err := j.cl.Issue.SearchWithContext(ctx, query, nil)
	if err != nil {
		return nil, fmt.Errorf("jira returned error: %w", err)
	}

	tickets := make([]task.Ticket, 0, len(issues))
	for _, issue := range issues {
		ticket := j.transformIssue(issue)
		if ticket.Parent, err = j.loadParent(issue); err != nil {
			return nil, fmt.Errorf("load parent: %w", err)
		}
		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

// Get returns a single task by its ID.
func (j *Jira) Get(ctx context.Context, key string) (task.Ticket, error) {
	issue, _, err := j.cl.Issue.GetWithContext(ctx, key, nil)
	if err != nil {
		return task.Ticket{}, fmt.Errorf("jira returned error: %w", err)
	}

	ticket := j.transformIssue(*issue)

	if ticket.Parent, err = j.loadParent(*issue); err != nil {
		return task.Ticket{}, fmt.Errorf("load parent: %w", err)
	}

	return ticket, nil
}

func (j *Jira) loadParent(issue jira.Issue) (*task.Ticket, error) {
	if issue.Fields.Parent == nil {
		return nil, nil
	}

	ticket, err := j.Get(context.Background(), issue.Fields.Parent.Key)
	if err != nil {
		return nil, fmt.Errorf("get parent ticket %s: %w", issue.Fields.Parent.Key, err)
	}

	return &ticket, nil
}

func (j *Jira) transformIssue(issue jira.Issue) task.Ticket {
	return task.Ticket{
		ID:       issue.Key,
		Name:     issue.Fields.Summary,
		Body:     issue.Fields.Description,
		ClosedAt: time.Time(issue.Fields.Resolutiondate),
		Author:   j.transformUser(issue.Fields.Creator),
		Assignee: j.transformUser(issue.Fields.Assignee),
	}
}

func (j *Jira) transformUser(user *jira.User) task.User {
	if user == nil {
		return task.User{}
	}

	return task.User{
		Username: user.Name,
		Email:    user.EmailAddress,
	}
}
