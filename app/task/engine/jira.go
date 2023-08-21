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
	"net/url"
	"strings"
	"time"
)

// Jira is a Jira task tracker engine.
type Jira struct {
	cl              *jira.Client
	baseURL         string
	epicFieldIDs    []string
	flaggedFieldIDs []string
}

// JiraParams is a set of parameters for Jira engine.
type JiraParams struct {
	URL        string
	Token      string
	HTTPClient http.Client
}

// NewJira creates a new Jira engine.
func NewJira(ctx context.Context, params JiraParams) (*Jira, error) {
	rq := requester.New(params.HTTPClient,
		middleware.Header("Authorization", "Bearer "+params.Token),
		logger.New(logger.Func(log.Printf), logger.Prefix("[DEBUG]")).Middleware,
	)

	cl, err := jira.NewClient(rq, params.URL)
	if err != nil {
		return nil, err
	}

	j := &Jira{cl: cl, baseURL: params.URL}

	ctx, cancel := context.WithTimeout(ctx, defaultSetupTimeout)
	defer cancel()

	if err = j.fillFieldIDs(ctx); err != nil {
		return nil, fmt.Errorf("get epic key: %w", err)
	}

	return j, nil
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

func (j *Jira) fillFieldIDs(ctx context.Context) error {
	fields, _, err := j.cl.Field.GetListWithContext(ctx)
	if err != nil {
		return fmt.Errorf("list jira fields: %w", err)
	}

	seek := func(key string) (result []string) {
		key = strings.ToLower(key)
		for _, field := range fields {
			if strings.ToLower(field.Name) == key {
				result = append(result, field.ID)
			}
		}
		return result
	}

	j.epicFieldIDs = seek("Epic Link")
	j.flaggedFieldIDs = seek("Flagged")

	return nil
}

var ticketTypeMapping = map[string]task.Type{
	"epic": task.TypeEpic,
	"bug":  task.TypeTask, "task": task.TypeTask, "story": task.TypeTask,
	"sub-task": task.TypeSubtask, "sub-story": task.TypeSubtask, "sub-bug": task.TypeSubtask,
}

func (j *Jira) transformIssue(issue jira.Issue) task.Ticket {
	u, _ := url.JoinPath(j.baseURL, "browse", issue.Key)

	ticket := task.Ticket{
		ID:       issue.Key,
		URL:      u,
		Name:     issue.Fields.Summary,
		Body:     issue.Fields.Description,
		ClosedAt: time.Time(issue.Fields.Resolutiondate),
		Author:   j.transformUser(issue.Fields.Creator),
		Assignee: j.transformUser(issue.Fields.Assignee),
		Type:     ticketTypeMapping[strings.ToLower(issue.Fields.Type.Name)],
		TypeRaw:  issue.Fields.Type.Name,
	}

	switch {
	case issue.Fields.Parent != nil:
		ticket.ParentID = issue.Fields.Parent.Key
	case issue.Fields.Epic != nil:
		ticket.ParentID = issue.Fields.Epic.Key
	default:
		if epicID, ok := j.seekCustomField(issue, j.epicFieldIDs).(string); ok {
			ticket.ParentID = epicID
		}
	}

	ticket.Flagged = j.seekCustomField(issue, j.flaggedFieldIDs) != nil

	return ticket
}

func (j *Jira) seekCustomField(issue jira.Issue, ids []string) interface{} {
	for _, fieldID := range ids {
		obj, ok := issue.Fields.Unknowns[fieldID]
		if !ok || obj == nil {
			continue
		}

		return obj
	}

	return nil
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
