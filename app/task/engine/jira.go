package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Semior001/releaseit/app/task"
	"github.com/andygrunwald/go-jira"
	"github.com/go-pkgz/requester"
	"github.com/go-pkgz/requester/middleware"
	"github.com/go-pkgz/requester/middleware/logger"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

// Jira is a Jira task tracker engine.
type Jira struct {
	httpCl          *http.Client
	cl              *jira.Client
	baseURL         string
	epicFieldIDs    []string
	flaggedFieldIDs []string

	JiraParams
}

// JiraParams is a set of parameters for Jira engine.
type JiraParams struct {
	BaseURL    string
	Token      string
	HTTPClient http.Client
	Enricher   struct {
		LoadWatchers bool
	}
}

// NewJira creates a new Jira engine.
func NewJira(ctx context.Context, params JiraParams) (*Jira, error) {
	rq := requester.New(params.HTTPClient,
		middleware.Header("Authorization", "Bearer "+params.Token),
		logger.New(logger.Func(log.Printf), logger.Prefix("[DEBUG]")).Middleware,
	)

	cl, err := jira.NewClient(rq, params.BaseURL)
	if err != nil {
		return nil, err
	}

	j := &Jira{cl: cl, baseURL: params.BaseURL, JiraParams: params, httpCl: rq.Client()}

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

	tickets := lo.Map(issues, func(item jira.Issue, _ int) task.Ticket {
		return j.transformIssue(item)
	})

	if tickets, err = j.enrich(ctx, tickets); err != nil {
		return nil, fmt.Errorf("enrich tickets: %w", err)
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

	tickets, err := j.enrich(ctx, []task.Ticket{ticket})
	if err != nil {
		return task.Ticket{}, fmt.Errorf("enrich tickets: %w", err)
	}

	return tickets[0], nil
}

func (j *Jira) enrich(ctx context.Context, tickets []task.Ticket) ([]task.Ticket, error) {
	if !j.Enricher.LoadWatchers {
		return tickets, nil
	}

	ticketsWithWatchers := lo.FilterMap(tickets, func(item task.Ticket, idx int) (int, bool) {
		return idx, item.WatchesCount > 0
	})

	if len(ticketsWithWatchers) == 0 {
		return tickets, nil
	}

	ewg, ctx := errgroup.WithContext(ctx)
	for _, ticketIdx := range ticketsWithWatchers {
		ticketIdx := ticketIdx
		ewg.Go(func() error {
			ticket := tickets[ticketIdx]
			defer func() { tickets[ticketIdx] = ticket }()

			watchers, err := j.listWatchers(ctx, ticket.ID)
			if err != nil {
				return fmt.Errorf("list watchers: %w", err)
			}
			ticket.Watchers = watchers

			return nil
		})
	}

	if err := ewg.Wait(); err != nil {
		return tickets, fmt.Errorf("enrich watchers: %w", err)
	}

	return tickets, nil
}

func (j *Jira) listWatchers(ctx context.Context, key string) ([]task.User, error) {
	u := fmt.Sprintf("%s/rest/api/2/issue/%s/watchers", j.baseURL, key)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := j.httpCl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	defer resp.Body.Close()

	var body struct {
		Watchers []*jira.User `json:"watchers"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return lo.Map(body.Watchers, func(item *jira.User, _ int) task.User {
		return j.transformUser(item)
	}), nil
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

	if issue.Fields.Watches != nil {
		ticket.WatchesCount = issue.Fields.Watches.WatchCount
	}

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
