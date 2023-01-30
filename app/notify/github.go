package notify

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/go-pkgz/requester"
	"github.com/go-pkgz/requester/middleware"
	gh "github.com/google/go-github/v37/github"
	"github.com/samber/lo"
)

// Github makes a new release on Github on the given version.
type Github struct {
	GithubParams

	cl              *gh.Client
	releaseNameTmpl *template.Template
}

// GithubParams describes parameters to initialize github releaser.
type GithubParams struct {
	Owner               string
	Name                string
	BasicAuthUsername   string
	BasicAuthPassword   string
	HTTPClient          http.Client
	ReleaseNameTmplText string
}

// NewGithub makes new instance of Github.
func NewGithub(params GithubParams) (*Github, error) {
	svc := &Github{GithubParams: params}

	cl := requester.New(params.HTTPClient)

	if params.BasicAuthUsername != "" && params.BasicAuthPassword != "" {
		cl.Use(middleware.BasicAuth(params.BasicAuthUsername, params.BasicAuthPassword))
	}

	svc.cl = gh.NewClient(cl.Client())

	_, _, err := svc.cl.Repositories.Get(context.Background(), svc.Owner, svc.Name)
	if err != nil {
		return nil, fmt.Errorf("check connection to github: %w", err)
	}

	svc.releaseNameTmpl, err = template.New("github_notify").
		Funcs(lo.OmitByKeys(sprig.FuncMap(), []string{"env", "expandenv"})).
		Parse(svc.ReleaseNameTmplText)
	if err != nil {
		return nil, fmt.Errorf("parse release name template: %w", err)
	}

	return svc, nil
}

// String returns the string representation of the destination.
func (g *Github) String() string {
	return fmt.Sprintf("github on %s/%s", g.Owner, g.Name)
}

type releaseNameTmplData struct {
	TagName string
}

// Send makes new release on github repository.
func (g *Github) Send(ctx context.Context, tagName, text string) error {
	if tagName == "" {
		return errors.New("tag name is empty")
	}

	buf := &strings.Builder{}

	err := g.releaseNameTmpl.Execute(buf, releaseNameTmplData{TagName: tagName})
	if err != nil {
		return fmt.Errorf("build release name: %w", err)
	}

	release := &gh.RepositoryRelease{
		TagName: &tagName,
		Name:    lo.ToPtr(buf.String()),
		Body:    &text,
	}

	if _, _, err = g.cl.Repositories.CreateRelease(ctx, g.Owner, g.Name, release); err != nil {
		return fmt.Errorf("github returned error: %w", err)
	}

	return nil
}
