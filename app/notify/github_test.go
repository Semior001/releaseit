package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gh "github.com/google/go-github/v37/github"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ghReleaseNameTmplText = `{{ .Tag.Name }} {{ .Tag.Message }} {{ .Tag.Author }} {{ .Tag.Date }}
{{ .Commit.SHA }} {{ .Commit.Message }} 
{{ .Commit.Author.Name }} {{ .Commit.Author.Date }} 
{{ .Commit.Committer.Name }} {{ .Commit.Committer.Date }}`

const ghExpectedReleaseName = `v1.0.0 message name 2020-01-01 00:00:00 +0000 UTC
sha commit message 
author-name 2020-01-01 00:00:00 +0000 UTC 
committer-name 2020-01-01 00:00:00 +0000 UTC`

func TestGithub_Send(t *testing.T) {
	t.Run("tag name is filled", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, pwd, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, "username", u)
			assert.Equal(t, "password", pwd)

			if r.URL.Path == "/repos/owner/name" {
				w.WriteHeader(http.StatusOK)
				return
			}

			if r.URL.Path == "/repos/owner/name/git/tags/v1.0.0" {
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(&gh.Tag{
					Tag:     gh.String("v1.0.0"),
					Message: gh.String("message"),
					Tagger: &gh.CommitAuthor{
						Name: gh.String("name"),
						Date: lo.ToPtr(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
					},
					Object: &gh.GitObject{
						SHA:  gh.String("sha"),
						Type: gh.String("commit"),
					},
				})
				require.NoError(t, err)
				return
			}

			if r.URL.Path == "/repos/owner/name/git/commits/sha" {
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(&gh.Commit{
					SHA:     gh.String("sha"),
					Message: gh.String("commit message"),
					Author: &gh.CommitAuthor{
						Name: gh.String("author-name"),
						Date: lo.ToPtr(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
					},
					Committer: &gh.CommitAuthor{
						Name: gh.String("committer-name"),
						Date: lo.ToPtr(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
					},
				})
				require.NoError(t, err)
				return
			}

			assert.Equal(t, "/repos/owner/name/releases", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var body gh.RepositoryRelease

			err := json.NewDecoder(r.Body).Decode(&body)
			require.NoError(t, err)

			assert.Equal(t, "v1.0.0", *body.TagName)
			assert.Equal(t, ghExpectedReleaseName, *body.Name)
			assert.Equal(t, "body", *body.Body)

			w.WriteHeader(http.StatusCreated)
		}))
		defer ts.Close()

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
			Tag:                 "v1.0.0",
			ReleaseNameTmplText: ghReleaseNameTmplText,
		})
		require.NoError(t, err)

		err = svc.Send(context.Background(), "body")
		require.NoError(t, err)
	})

}

func TestGithub_String(t *testing.T) {
	assert.Equal(t, "github on owner/name", (&Github{GithubParams: GithubParams{Name: "name", Owner: "owner"}}).String())
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
