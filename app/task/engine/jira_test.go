package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Semior001/releaseit/app/task"
	"github.com/andygrunwald/go-jira"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type J = map[string]interface{}

func TestJira_List(t *testing.T) {
	j := newJira(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/rest/api/2/search", r.URL.Path, "path is not set")
		require.Equal(t, http.MethodGet, r.Method, "method is not set")

		jql := r.URL.Query().Get("jql")
		assert.Equal(t, "key in (KEY-1,KEY-2)", jql, "jql is not set")

		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(J{"issues": []jira.Issue{
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
		}})
		require.NoError(t, err)
	})

	tickets, err := j.List(context.Background(), []string{"KEY-1", "KEY-2"})
	require.NoError(t, err)

	assert.Equal(t, []task.Ticket{
		{
			ID:       "KEY-1",
			URL:      fmt.Sprintf("%s/browse/KEY-1", j.baseURL),
			Name:     "summary",
			Body:     "description",
			ClosedAt: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Author:   task.User{Username: "creator", Email: "creator@jira.com"},
			Assignee: task.User{Username: "assignee", Email: "assignee@jira.com"},
		},
		{
			ID:       "KEY-2",
			URL:      fmt.Sprintf("%s/browse/KEY-2", j.baseURL),
			ParentID: "KEY-3",
			Name:     "summary-1",
			Body:     "description-1",
			ClosedAt: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC),
			Author:   task.User{Username: "creator1", Email: "creator1@jira.com"},
		},
	}, utcTimes(tickets))
}

func TestJira_Get(t *testing.T) {
	t.Run("parent explicitly defined", func(t *testing.T) {
		j := newJira(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/rest/api/2/issue/KEY-2", r.URL.Path, "path is not set")
			require.Equal(t, http.MethodGet, r.Method, "method is not set")

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
		})

		ticket, err := j.Get(context.Background(), "KEY-2")
		require.NoError(t, err)

		assert.Equal(t, []task.Ticket{{
			ID:       "KEY-2",
			URL:      fmt.Sprintf("%s/browse/KEY-2", j.baseURL),
			ParentID: "KEY-3",
			Name:     "summary-1",
			Body:     "description-1",
			ClosedAt: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC),
			Author:   task.User{Username: "creator1", Email: "creator1@jira.com"},
		}}, utcTimes([]task.Ticket{ticket}))
	})

	t.Run("parent is epic in unknown field", func(t *testing.T) {
		j := newJira(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/rest/api/2/issue/KEY-2", r.URL.Path, "path is not set")
			require.Equal(t, http.MethodGet, r.Method, "method is not set")

			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(jira.Issue{
				Key: "KEY-2",
				Fields: &jira.IssueFields{
					Summary:        "summary-1",
					Description:    "description-1",
					Resolutiondate: jira.Time(time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)),
					Creator:        &jira.User{Name: "creator1", EmailAddress: "creator1@jira.com"},
					Unknowns: J{
						"customfield_10002": "KEY-3",
						"customfield_10001": true, // value doesn't matter, only it's presence
					},
				},
			})
			require.NoError(t, err)
		})

		ticket, err := j.Get(context.Background(), "KEY-2")
		require.NoError(t, err)

		assert.Equal(t, []task.Ticket{{
			ID:       "KEY-2",
			URL:      fmt.Sprintf("%s/browse/KEY-2", j.baseURL),
			ParentID: "KEY-3",
			Name:     "summary-1",
			Body:     "description-1",
			ClosedAt: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC),
			Author:   task.User{Username: "creator1", Email: "creator1@jira.com"},
			Flagged:  true,
		}}, utcTimes([]task.Ticket{ticket}))
	})

	t.Run("epic is explicitly defined", func(t *testing.T) {
		j := newJira(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, "/rest/api/2/issue/KEY-2", r.URL.Path, "path is not set")
			require.Equal(t, http.MethodGet, r.Method, "method is not set")

			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(jira.Issue{
				Key:  "KEY-2",
				Self: "https://some-jira-instance/KEY-2",
				Fields: &jira.IssueFields{
					Summary:        "summary-1",
					Description:    "description-1",
					Resolutiondate: jira.Time(time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)),
					Creator:        &jira.User{Name: "creator1", EmailAddress: "creator1@jira.com"},
					Epic:           &jira.Epic{Key: "KEY-3"},
					Unknowns: J{
						"customfield_10002": "KEY-199",
						"customfield_10001": true, // value doesn't matter, only it's presence
					},
				},
			})
			require.NoError(t, err)
		})

		ticket, err := j.Get(context.Background(), "KEY-2")
		require.NoError(t, err)

		assert.Equal(t, []task.Ticket{{
			ID:       "KEY-2",
			URL:      fmt.Sprintf("%s/browse/KEY-2", j.baseURL),
			ParentID: "KEY-3",
			Name:     "summary-1",
			Body:     "description-1",
			ClosedAt: time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC),
			Author:   task.User{Username: "creator1", Email: "creator1@jira.com"},
			Flagged:  true,
		}}, utcTimes([]task.Ticket{ticket}))
	})

	t.Run("watchers were requested", func(t *testing.T) {
		j := newJira(t, func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, http.MethodGet, r.Method, "method is not set")
			switch r.URL.Path {
			case "/rest/api/2/issue/KEY-2":
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(jira.Issue{
					Key:  "KEY-2",
					Self: "https://some-jira-instance/KEY-2",
					Fields: &jira.IssueFields{
						Summary:        "summary-1",
						Description:    "description-1",
						Resolutiondate: jira.Time(time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)),
						Creator:        &jira.User{Name: "creator1", EmailAddress: "creator1@jira.com"},
						Epic:           &jira.Epic{Key: "KEY-3"},
						Watches:        &jira.Watches{WatchCount: 1},
						Unknowns: J{
							"customfield_10002": "KEY-199",
							"customfield_10001": true, // value doesn't matter, only it's presence
						},
					},
				})
				require.NoError(t, err)
			case "/rest/api/2/issue/KEY-2/watchers":
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(J{
					"self":       "https://some-jira-instance/KEY-2/watchers",
					"isWatching": true,
					"watchCount": 1,
					"watchers":   []J{{"name": "watcher1", "emailAddress": "example@test.com"}},
				})
				require.NoError(t, err)
			default:
				require.Fail(t, "unexpected path", r.URL.Path)
			}
		})

		ticket, err := j.Get(context.Background(), "KEY-2")
		require.NoError(t, err)

		assert.Equal(t, []task.Ticket{{
			ID:           "KEY-2",
			URL:          fmt.Sprintf("%s/browse/KEY-2", j.baseURL),
			ParentID:     "KEY-3",
			Name:         "summary-1",
			Body:         "description-1",
			ClosedAt:     time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC),
			Author:       task.User{Username: "creator1", Email: "creator1@jira.com"},
			Flagged:      true,
			WatchesCount: 1,
			Watchers:     []task.User{{Username: "watcher1", Email: "example@test.com"}},
		}}, utcTimes([]task.Ticket{ticket}))
	})
}

func newJira(t *testing.T, h http.HandlerFunc) *Jira {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer abacaba", r.Header.Get("Authorization"))

		if r.URL.Path == "/rest/api/2/field" {
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode([]jira.Field{
				{ID: "customfield_10000", Name: "epic link"},
				{ID: "customfield_10001", Name: "flagged"},
				{ID: "customfield_10002", Name: "epic link"},
			})
			require.NoError(t, err)
			return
		}

		h(w, r)
	}))
	t.Cleanup(ts.Close)

	params := JiraParams{
		BaseURL:    ts.URL,
		Token:      "abacaba",
		HTTPClient: *ts.Client(),
	}
	params.Enricher.LoadWatchers = true

	svc, err := NewJira(context.Background(), params)
	require.NoError(t, err)

	return svc
}

func utcTimes(tickets []task.Ticket) []task.Ticket {
	for i := range tickets {
		tickets[i].ClosedAt = tickets[i].ClosedAt.UTC()
	}

	return tickets
}
