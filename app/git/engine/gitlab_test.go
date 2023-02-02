package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Semior001/releaseit/app/git"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gl "github.com/xanzy/go-gitlab"
)

func TestGitlab_Compare(t *testing.T) {
	svc := newGitlab(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v4/projects/projectID/repository/compare", r.URL.Path)

		q := r.URL.Query()
		assert.Equal(t, "old", q.Get("from"))
		assert.Equal(t, "new", q.Get("to"))

		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode(gl.Compare{Commits: []*gl.Commit{
			{
				ID:        "sha",
				Message:   "message",
				ParentIDs: []string{"parent"},
				Stats:     &gl.CommitStats{Total: 1, Additions: 1, Deletions: 2},
			},
			{
				ID:        "sha2",
				Message:   "message2",
				ParentIDs: []string{"parent2", "parent3"},
				Stats:     &gl.CommitStats{Total: 1, Additions: 1, Deletions: 2},
			},
		}})
		require.NoError(t, err)
	})

	cmp, err := svc.Compare(context.Background(), "old", "new")
	require.NoError(t, err)

	assert.Equal(t, git.CommitsComparison{
		TotalCommits: 2,
		Commits: []git.Commit{
			{
				SHA:         "sha",
				Message:     "message",
				ParentSHAs:  []string{"parent"},
				CommitStats: git.CommitStats{Total: 1, Additions: 1, Deletions: 2},
			},
			{
				SHA:         "sha2",
				Message:     "message2",
				ParentSHAs:  []string{"parent2", "parent3"},
				CommitStats: git.CommitStats{Total: 1, Additions: 1, Deletions: 2},
			},
		},
	}, cmp)
}

func TestGitlab_GetLastCommitOfBranch(t *testing.T) {
	svc := newGitlab(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v4/projects/projectID/repository/branches/branch", r.URL.Path)

		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode(gl.Branch{Commit: &gl.Commit{ID: "sha"}})
		require.NoError(t, err)
	})

	sha, err := svc.GetLastCommitOfBranch(context.Background(), "branch")
	require.NoError(t, err)
	assert.Equal(t, "sha", sha)
}

func TestGitlab_ListPRsOfCommit(t *testing.T) {
	now := time.Now()

	svc := newGitlab(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v4/projects/projectID/repository/commits/sha/merge_requests", r.URL.Path)

		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode([]*gl.MergeRequest{{
			IID:          1,
			Title:        "title",
			Description:  "description",
			Author:       &gl.BasicUser{Username: "author"},
			Labels:       []string{"label1", "label2"},
			ClosedAt:     lo.ToPtr(now.UTC()),
			SourceBranch: "source",
			WebURL:       "url",
		}})
		require.NoError(t, err)
	})

	prs, err := svc.ListPRsOfCommit(context.Background(), "sha")
	require.NoError(t, err)
	assert.Equal(t, []git.PullRequest{{
		Number:   1,
		Title:    "title",
		Body:     "description",
		Author:   git.User{Username: "author"},
		Labels:   []string{"label1", "label2"},
		ClosedAt: now.UTC(),
		Branch:   "source",
		URL:      "url",
	}}, prs)
}

func TestGitlab_ListTags(t *testing.T) {
	svc := newGitlab(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v4/projects/projectID/repository/tags", r.URL.Path)

		w.WriteHeader(http.StatusOK)

		err := json.NewEncoder(w).Encode([]*gl.Tag{{
			Name: "v1.0.0",
			Commit: &gl.Commit{
				ID:        "sha",
				ParentIDs: []string{"parent"},
				Message:   "message",
				Stats:     &gl.CommitStats{Total: 1, Additions: 1, Deletions: 2},
			},
		}})
		require.NoError(t, err)
	})

	tags, err := svc.ListTags(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []git.Tag{{
		Name: "v1.0.0",
		Commit: git.Commit{
			SHA:         "sha",
			ParentSHAs:  []string{"parent"},
			Message:     "message",
			CommitStats: git.CommitStats{Total: 1, Additions: 1, Deletions: 2},
		},
	}}, tags)
}

func newGitlab(t *testing.T, h http.HandlerFunc) *Gitlab {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("PRIVATE-TOKEN") != "token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.URL.Path == "/api/v4/" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.URL.Path == "/api/v4/projects/projectID" {
			w.WriteHeader(http.StatusOK)

			err := json.NewEncoder(w).Encode(gl.Project{ID: 1})
			require.NoError(t, err)

			return
		}

		h(w, r)
	}))

	svc, err := NewGitlab("token", ts.URL, "projectID", http.Client{
		Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
			return http.DefaultTransport.RoundTrip(req)
		}),
	})
	require.NoError(t, err)

	return svc
}
