package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gh "github.com/google/go-github/v37/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGithub_Send(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, pwd, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "username", u)
		assert.Equal(t, "password", pwd)

		if r.URL.Path == "/repos/owner/name" {
			w.WriteHeader(http.StatusOK)
			return
		}

		assert.Equal(t, "/repos/owner/name/releases", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body gh.RepositoryRelease

		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)

		assert.Equal(t, "tag", *body.TagName)
		assert.Equal(t, "release name", *body.Name)
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
		ReleaseNameTmplText: "release name",
	})
	require.NoError(t, err)

	err = svc.Send(context.Background(), "tag", "body")
	require.NoError(t, err)
}

type roundTripperFunc func(req *http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
