package notify

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Semior001/releaseit/app/service/eval"
	"github.com/go-pkgz/requester"
	"github.com/go-pkgz/requester/middleware"
	gh "github.com/google/go-github/v37/github"
	"github.com/samber/lo"
)

// Github makes a new release on Github on the given version.
type Github struct {
	GithubParams

	cl *gh.Client
}

// GithubParams describes parameters to initialize github releaser.
type GithubParams struct {
	Evaluator           *eval.Evaluator
	Owner               string
	Name                string
	BasicAuthUsername   string
	BasicAuthPassword   string
	HTTPClient          http.Client
	ReleaseNameTmplText string
	Tag                 string
	Extras              map[string]string
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

	return svc, nil
}

// String returns the string representation of the destination.
func (g *Github) String() string {
	return fmt.Sprintf("github on %s/%s", g.Owner, g.Name)
}

type releaseNameTmplData struct {
	Extras map[string]string
	Tag    struct {
		Name    string
		Message string
		Author  string
		Date    time.Time
	}
	Commit struct {
		SHA     string
		Message string
		Author  struct {
			Name string
			Date time.Time
		}
		Committer struct {
			Name string
			Date time.Time
		}
	}
}

// Send makes new release on github repository.
func (g *Github) Send(ctx context.Context, text string) error {
	// get tag message
	tag, _, err := g.cl.Git.GetTag(ctx, g.Owner, g.Name, g.Tag)
	if err != nil {
		return fmt.Errorf("get tag %s: %w", g.Tag, err)
	}

	data := releaseNameTmplData{}
	data.Tag.Name = g.Tag
	data.Tag.Message = tag.GetMessage()
	data.Tag.Author = tag.GetTagger().GetName()
	data.Tag.Date = tag.GetTagger().GetDate()
	data.Commit.SHA = tag.GetObject().GetSHA()
	data.Extras = g.Extras

	if tag.GetObject().GetType() == "commit" {
		cmt, _, err := g.cl.Git.GetCommit(ctx, g.Owner, g.Name, tag.GetObject().GetSHA())
		if err != nil {
			return fmt.Errorf("get commit %s: %w", tag.GetObject().GetSHA(), err)
		}

		data.Commit.Message = cmt.GetMessage()
		data.Commit.Author.Name = cmt.GetAuthor().GetName()
		data.Commit.Author.Date = cmt.GetAuthor().GetDate()
		data.Commit.Committer.Name = cmt.GetCommitter().GetName()
		data.Commit.Committer.Date = cmt.GetCommitter().GetDate()
	}

	name, err := g.Evaluator.Evaluate(ctx, g.ReleaseNameTmplText, data)
	if err != nil {
		return fmt.Errorf("build release name: %w", err)
	}
	release := &gh.RepositoryRelease{
		TagName: tag.Tag,
		Name:    lo.ToPtr(name),
		Body:    &text,
	}

	if _, _, err = g.cl.Repositories.CreateRelease(ctx, g.Owner, g.Name, release); err != nil {
		return fmt.Errorf("github returned error: %w", err)
	}

	return nil
}
