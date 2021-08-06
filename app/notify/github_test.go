package notify

import (
	"context"
	"encoding/json"
	"net/http"
	"sync/atomic"
	"testing"
	"text/template"

	"github.com/Semior001/releaseit/app/store"
	"github.com/Semior001/releaseit/app/util"
	gh "github.com/google/go-github/v37/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGithub_Send(t *testing.T) {
	changelog := store.Changelog{
		Tag: store.Tag{Name: "Awesome tag"},
		ClosedPRs: []store.PullRequest{{
			Number: 1,
			Title:  "first PR",
			Body:   "PR body",
		}},
	}

	var hit int32
	cl := util.RedirectingTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		req := struct {
			Name string `json:"name"`
			Body string `json:"body"`
		}{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		assert.Equal(t, req.Name, "Version: Awesome tag")
		assert.Equal(t, req.Body, "some awesome\nmultilined changelog")
		atomic.AddInt32(&hit, 1)
	})

	svc := &Github{
		GithubParams: GithubParams{
			Owner:             "testowner",
			Name:              "testname",
			BasicAuthUsername: "Semior001",
			BasicAuthPassword: "someawesomepassword",
			ReleaseNotesBuilder: &releaseBuilderMock{BuildFunc: func(cl store.Changelog) (string, error) {
				assert.Equal(t, changelog, cl)
				return "some awesome\nmultilined changelog", nil
			}},
		},
		cl:              gh.NewClient(cl),
		releaseNameTmpl: mustParseTmpl(t, "Version: {{.TagName}}"),
	}

	err := svc.Send(context.Background(), changelog)
	require.NoError(t, err)

	assert.Equal(t, int32(1), atomic.LoadInt32(&hit))
}

func TestGithub_String(t *testing.T) {
	assert.Equal(t, "github on Semior001/releaseit",
		(&Github{GithubParams: GithubParams{Owner: "Semior001", Name: "releaseit"}}).String())
}

func mustParseTmpl(t *testing.T, s string) *template.Template {
	tmpl, err := template.New("github_notify_name").Parse(s)
	require.NoError(t, err)
	return tmpl
}
