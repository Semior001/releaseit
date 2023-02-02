package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Semior001/releaseit/app/git"
	gh "github.com/google/go-github/v37/github"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGithub_GetLastCommitOfBranch(t *testing.T) {
	svc := newGithub(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/repos/owner/name/branches/branch", r.URL.Path, "path is not set")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"commit": {"sha": "sha"}}`))
		require.NoError(t, err)
	})

	sha, err := svc.GetLastCommitOfBranch(context.Background(), "branch")
	require.NoError(t, err)
	require.Equal(t, "sha", sha)
}

func TestGithub_Compare(t *testing.T) {
	svc := newGithub(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/repos/owner/name/compare/old...new", r.URL.Path, "path is not set")

		err := json.NewEncoder(w).Encode(gh.CommitsComparison{
			Commits: []*gh.RepositoryCommit{
				{
					SHA:     gh.String("sha"),
					Commit:  &gh.Commit{Message: gh.String("message")},
					Parents: []*gh.Commit{{SHA: gh.String("parent")}},
				},
				{
					SHA:     gh.String("sha2"),
					Commit:  &gh.Commit{Message: gh.String("message2")},
					Parents: []*gh.Commit{{SHA: gh.String("parent2")}, {SHA: gh.String("parent3")}},
				},
			},
			TotalCommits: gh.Int(2),
		})
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	})

	comp, err := svc.Compare(context.Background(), "old", "new")
	require.NoError(t, err)

	assert.Equal(t, git.CommitsComparison{
		Commits: []git.Commit{
			{SHA: "sha", ParentSHAs: []string{"parent"}, Message: "message"},
			{SHA: "sha2", ParentSHAs: []string{"parent2", "parent3"}, Message: "message2"},
		},
		TotalCommits: 2,
	}, comp)
}

func TestGithub_ListPRsOfCommit(t *testing.T) {
	now := time.Now()

	svc := newGithub(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/repos/owner/name/commits/sha/pulls", r.URL.Path, "path is not set")

		err := json.NewEncoder(w).Encode([]*gh.PullRequest{
			{
				Number:   gh.Int(1),
				Title:    gh.String("title 1"),
				Body:     gh.String("body 1"),
				ClosedAt: lo.ToPtr(now.UTC()),
				User:     &gh.User{Login: gh.String("username 1"), Email: gh.String("email 1")},
				Labels:   []*gh.Label{{Name: gh.String("label 1")}},
				Base:     &gh.PullRequestBranch{Ref: gh.String("branch 1")},
				URL:      gh.String("url 1"),
			},
			{
				Number:   gh.Int(2),
				Title:    gh.String("title 2"),
				Body:     gh.String("body 2"),
				ClosedAt: lo.ToPtr(now.Add(time.Hour).UTC()),
				User:     &gh.User{Login: gh.String("username 2"), Email: gh.String("email 2")},
				Labels:   []*gh.Label{{Name: gh.String("label 2")}, {Name: gh.String("label 3")}},
				Base:     &gh.PullRequestBranch{Ref: gh.String("branch 2")},
				URL:      gh.String("url 2"),
			},
		})
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	})

	prs, err := svc.ListPRsOfCommit(context.Background(), "sha")
	require.NoError(t, err)
	assert.Equal(t, []git.PullRequest{
		{
			Number:   1,
			Title:    "title 1",
			Body:     "body 1",
			Author:   git.User{Username: "username 1", Email: "email 1"},
			Labels:   []string{"label 1"},
			ClosedAt: now.UTC(),
			Branch:   "branch 1",
			URL:      "url 1",
		},
		{
			Number:   2,
			Title:    "title 2",
			Body:     "body 2",
			Author:   git.User{Username: "username 2", Email: "email 2"},
			Labels:   []string{"label 2", "label 3"},
			ClosedAt: now.Add(time.Hour).UTC(),
			Branch:   "branch 2",
			URL:      "url 2",
		},
	}, prs)
}

func TestGithub_ListTags(t *testing.T) {
	svc := newGithub(t, func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/repos/owner/name/tags", r.URL.Path, "path is not set")

		err := json.NewEncoder(w).Encode([]*gh.RepositoryTag{
			{
				Name: gh.String("v0.1.0"),
				Commit: &gh.Commit{
					SHA:     gh.String("sha 1"),
					Parents: []*gh.Commit{{SHA: gh.String("parent 1")}},
					Message: gh.String("message 1"),
				},
			},
			{
				Name: gh.String("v0.2.0"),
				Commit: &gh.Commit{
					SHA: gh.String("sha 2"),
					Parents: []*gh.Commit{
						{SHA: gh.String("parent 2")},
						{SHA: gh.String("parent 3")},
					},
					Message: gh.String("message 2"),
				},
			},
		})
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	})

	tags, err := svc.ListTags(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []git.Tag{
		{
			Name: "v0.1.0",
			Commit: git.Commit{
				SHA:        "sha 1",
				ParentSHAs: []string{"parent 1"},
				Message:    "message 1",
			},
		},
		{
			Name: "v0.2.0",
			Commit: git.Commit{
				SHA:        "sha 2",
				ParentSHAs: []string{"parent 2", "parent 3"},
				Message:    "message 2",
			},
		},
	}, tags)
}

func TestGithub_commitToStore(t *testing.T) {
	tests := []struct {
		name   string
		commit shaGetter
		want   git.Commit
	}{
		{
			name: "regular commit",
			commit: &gh.Commit{
				SHA:     gh.String("sha"),
				Stats:   &gh.CommitStats{Total: gh.Int(1), Additions: gh.Int(2), Deletions: gh.Int(3)},
				Parents: []*gh.Commit{{SHA: gh.String("parent1")}, {SHA: gh.String("parent2")}},
				Message: gh.String("message"),
			},
			want: git.Commit{
				SHA:         "sha",
				CommitStats: git.CommitStats{Total: 1, Additions: 2, Deletions: 3},
				ParentSHAs:  []string{"parent1", "parent2"},
				Message:     "message",
			},
		},
		{
			name: "repository commit",
			commit: &gh.RepositoryCommit{
				SHA:     gh.String("sha"),
				Stats:   &gh.CommitStats{Total: gh.Int(1), Additions: gh.Int(2), Deletions: gh.Int(3)},
				Parents: []*gh.Commit{{SHA: gh.String("parent1")}, {SHA: gh.String("parent2")}},
				Commit:  &gh.Commit{Message: gh.String("message")},
			},
			want: git.Commit{
				SHA:         "sha",
				CommitStats: git.CommitStats{Total: 1, Additions: 2, Deletions: 3},
				ParentSHAs:  []string{"parent1", "parent2"},
				Message:     "message",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, (*Github)(nil).commitToStore(tt.commit))
		})
	}
}

func newGithub(t *testing.T, h http.HandlerFunc) *Github {
	t.Helper()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, pwd, ok := r.BasicAuth()
		require.True(t, ok, "basic auth is not set")
		require.Equal(t, "username", u, "username is not set")
		require.Equal(t, "password", pwd, "password is not set")

		if r.URL.Path == "/repos/owner/name" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h(w, r)
	}))
	t.Cleanup(ts.Close)

	svc, err := NewGithub(GithubParams{
		Owner:             "owner",
		Name:              "name",
		BasicAuthUsername: "username",
		BasicAuthPassword: "password",
		HTTPClient: http.Client{
			Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				// hijack the request to test server
				req.URL.Host = ts.URL[7:]
				req.URL.Scheme = "http"
				return http.DefaultTransport.RoundTrip(req)
			}),
		},
	})
	require.NoError(t, err)

	return svc
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
