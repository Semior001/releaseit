package engine

import (
	"context"
	"encoding/json"
	"github.com/Semior001/releaseit/app/task"
	"github.com/andygrunwald/go-jira"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestJira_List(t *testing.T) {
	j := newJira(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/rest/api/2/issue/KEY-3" {
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(jira.Issue{
				Key: "KEY-3",
				Fields: &jira.IssueFields{
					Summary:        "summary-2",
					Description:    "description-2",
					Resolutiondate: jira.Time(time.Date(2020, 1, 1, 2, 0, 0, 0, time.UTC)),
					Creator:        &jira.User{Name: "creator2", EmailAddress: "creator2@jira.com"},
				},
			})
			require.NoError(t, err)
			return
		}

		require.Equal(t, "/rest/api/2/search", r.URL.Path, "path is not set")

		jql := r.URL.Query().Get("jql")
		assert.Equal(t, "key in (KEY-1,KEY-2)", jql, "jql is not set")

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(map[string]any{
			"issues": []jira.Issue{
				{
					Key: "KEY-1",
					Fields: &jira.IssueFields{
						Summary:        "summary",
						Description:    "description",
						Resolutiondate: jira.Time(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
						Creator:        &jira.User{Name: "creator", EmailAddress: "creator@jira.com"},
						Assignee:       &jira.User{Name: "assignee", EmailAddress: "assignee@jira.com"},
					},
				},
				{
					Key: "KEY-2",
					Fields: &jira.IssueFields{
						Summary:        "summary-1",
						Description:    "description-1",
						Resolutiondate: jira.Time(time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)),
						Creator:        &jira.User{Name: "creator1", EmailAddress: "creator1@jira.com"},
						Parent:         &jira.Parent{Key: "KEY-3"},
					},
				},
			},
		})
		require.NoError(t, err)
	})

	tickets, err := j.List(context.Background(), []string{"KEY-1", "KEY-2"})
	require.NoError(t, err)

	assert.Equal(t, []task.Ticket{
		{
			ID:       "KEY-1",
			Name:     "summary",
			Body:     "description",
			ClosedAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Author:   task.User{Username: "creator", Email: "creator@jira.com"},
			Assignee: task.User{Username: "assignee", Email: "assignee@jira.com"},
		},
		{
			ID:       "KEY-2",
			Name:     "summary-1",
			Body:     "description-1",
			ClosedAt: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC),
			Author:   task.User{Username: "creator1", Email: "creator1@jira.com"},
			Parent: &task.Ticket{
				ID:       "KEY-3",
				Name:     "summary-2",
				Body:     "description-2",
				ClosedAt: time.Date(2020, 1, 1, 2, 0, 0, 0, time.UTC),
				Author:   task.User{Username: "creator2", Email: "creator2@jira.com"},
			},
		},
	}, utcTimes(tickets))
}

func TestJira_Get(t *testing.T) {
	j := newJira(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rest/api/2/issue/KEY-2":
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(jira.Issue{
				Key: "KEY-2",
				Fields: &jira.IssueFields{
					Summary:        "summary-1",
					Description:    "description-1",
					Resolutiondate: jira.Time(time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)),
					Creator:        &jira.User{Name: "creator1", EmailAddress: "creator1@jira.com"},
					Parent:         &jira.Parent{Key: "KEY-3"},
				},
			})
			require.NoError(t, err)
		case "/rest/api/2/issue/KEY-3":
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(jira.Issue{
				Key: "KEY-3",
				Fields: &jira.IssueFields{
					Summary:        "summary-2",
					Description:    "description-2",
					Resolutiondate: jira.Time(time.Date(2020, 1, 1, 2, 0, 0, 0, time.UTC)),
					Creator:        &jira.User{Name: "creator2", EmailAddress: "creator2@jira.com"},
				},
			})
			require.NoError(t, err)
		default:
			require.Fail(t, "unexpected path: %s", r.URL.Path)
		}
	})

	ticket, err := j.Get(context.Background(), "KEY-2")
	require.NoError(t, err)

	assert.Equal(t, []task.Ticket{{
		ID:       "KEY-2",
		Name:     "summary-1",
		Body:     "description-1",
		ClosedAt: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC),
		Author:   task.User{Username: "creator1", Email: "creator1@jira.com"},
		Parent: &task.Ticket{
			ID:       "KEY-3",
			Name:     "summary-2",
			Body:     "description-2",
			ClosedAt: time.Date(2020, 1, 1, 2, 0, 0, 0, time.UTC),
			Author:   task.User{Username: "creator2", Email: "creator2@jira.com"},
		},
	}}, utcTimes([]task.Ticket{ticket}))
}

func newJira(t *testing.T, h http.HandlerFunc) *Jira {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer abacaba", r.Header.Get("Authorization"))
		h(w, r)
	}))
	t.Cleanup(ts.Close)

	svc, err := NewJira(JiraParams{
		URL:        ts.URL,
		Token:      "abacaba",
		HTTPClient: http.Client{},
	})
	require.NoError(t, err)

	return svc
}

func utcTimes(tickets []task.Ticket) []task.Ticket {
	for i := range tickets {
		tickets[i].ClosedAt = tickets[i].ClosedAt.UTC()
		if tickets[i].Parent != nil {
			tickets[i].Parent.ClosedAt = tickets[i].Parent.ClosedAt.UTC()
		}
	}

	return tickets
}
