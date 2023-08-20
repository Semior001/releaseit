package engine

import (
	"context"
	"fmt"
	"github.com/Semior001/releaseit/app/task"
	"github.com/andygrunwald/go-jira"
	"github.com/go-pkgz/requester"
	"github.com/go-pkgz/requester/middleware"
	"github.com/go-pkgz/requester/middleware/logger"
	"github.com/samber/lo"
	"log"
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
		logger.New(logger.Func(log.Printf), logger.Prefix("[DEBUG] ")).Middleware,
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

	return lo.Map(issues, func(item jira.Issue, _ int) task.Ticket {
		return j.transformIssue(item)
	}), nil
}

// Get returns a single task by its ID.
func (j *Jira) Get(ctx context.Context, key string) (task.Ticket, error) {
	issue, _, err := j.cl.Issue.GetWithContext(ctx, key, nil)
	if err != nil {
		return task.Ticket{}, fmt.Errorf("jira returned error: %w", err)
	}

	ticket := j.transformIssue(*issue)

	return ticket, nil
}

func (j *Jira) transformIssue(issue jira.Issue) task.Ticket {
	ticket := task.Ticket{
		ID:       issue.Key,
		Name:     issue.Fields.Summary,
		Body:     issue.Fields.Description,
		ClosedAt: time.Time(issue.Fields.Resolutiondate),
		Author:   j.transformUser(issue.Fields.Creator),
		Assignee: j.transformUser(issue.Fields.Assignee),
	}

	if issue.Fields.Parent != nil {
		ticket.ParentID = issue.Fields.Parent.Key
	}

	return ticket
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
