package engine

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Semior001/releaseit/app/git"
	"github.com/go-pkgz/requester"
	"github.com/go-pkgz/requester/middleware"
	gh "github.com/google/go-github/v37/github"
)

// Github implements Interface with github API below it.
type Github struct {
	cl    *gh.Client
	owner string
	name  string
}

// NewGithub makes new instance of Github.
func NewGithub(owner, name, basicAuthUsername, basicAuthPassword string, httpCl http.Client) (*Github, error) {
	svc := &Github{
		owner: owner,
		name:  name,
	}

	cl := requester.New(httpCl)

	if basicAuthUsername != "" && basicAuthPassword != "" {
		cl.Use(middleware.BasicAuth(basicAuthUsername, basicAuthPassword))
	}

	svc.cl = gh.NewClient(cl.Client())

	if _, _, err := svc.cl.Repositories.Get(context.Background(), svc.owner, svc.name); err != nil {
		return nil, fmt.Errorf("check connection to github: %w", err)
	}

	return svc, nil
}

// Compare two commits by their SHA.
func (g *Github) Compare(ctx context.Context, fromSHA, toSHA string) (git.CommitsComparison, error) {
	comp, _, err := g.cl.Repositories.CompareCommits(ctx, g.owner, g.name, fromSHA, toSHA)
	if err != nil {
		return git.CommitsComparison{}, fmt.Errorf("github returned error: %w", err)
	}

	commits := make([]git.Commit, len(comp.Commits))

	for i, commit := range comp.Commits {
		commits[i] = g.commitToStore(commit)
	}

	return git.CommitsComparison{
		Commits:      commits,
		TotalCommits: comp.GetTotalCommits(),
	}, nil
}

// ListPRsOfCommit returns pull requests associated with commit by the given SHA.
func (g *Github) ListPRsOfCommit(ctx context.Context, sha string) ([]git.PullRequest, error) {
	prs, _, err := g.cl.PullRequests.ListPullRequestsWithCommit(ctx, g.owner, g.name, sha, &gh.PullRequestListOptions{})
	if err != nil {
		return nil, fmt.Errorf("github returned error: %w", err)
	}

	res := make([]git.PullRequest, len(prs))

	for i, pr := range prs {
		res[i] = git.PullRequest{
			Number:   pr.GetNumber(),
			Title:    pr.GetTitle(),
			Body:     pr.GetBody(),
			ClosedAt: pr.GetClosedAt(),
			Author:   git.User{Username: pr.GetUser().GetLogin(), Email: pr.GetUser().GetEmail()},
			Labels:   make([]string, len(pr.Labels)),
		}

		for j, lbl := range pr.Labels {
			res[i].Labels[j] = lbl.GetName()
		}
	}

	return res, nil
}

// HeadCommit returns the alias of the oldest commit in the repository
func (g *Github) HeadCommit(_ context.Context) (string, error) {
	return "HEAD", nil
}

// ListTags returns all tags of the repository.
func (g *Github) ListTags(ctx context.Context) ([]git.Tag, error) {
	tags, _, err := g.cl.Repositories.ListTags(ctx, g.owner, g.name, &gh.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("github returned error: %w", err)
	}

	res := make([]git.Tag, len(tags))

	for i, tag := range tags {
		res[i] = git.Tag{
			Name:   tag.GetName(),
			Commit: g.commitToStore(tag.GetCommit()),
		}
	}

	return res, nil
}

type shaGetter interface {
	GetSHA() string
}

func (g *Github) commitToStore(commitInterface shaGetter) git.Commit {
	res := git.Commit{SHA: commitInterface.GetSHA()}
	switch cmt := commitInterface.(type) {
	case *gh.Commit:
		res.Author = g.commitAuthorToStore(cmt.GetAuthor())
		res.Committer = g.commitAuthorToStore(cmt.GetCommitter())
		res.ParentSHAs = transform(cmt.Parents, func(c *gh.Commit) string { return c.GetSHA() })
	case *gh.RepositoryCommit:
		res.Author = g.commitAuthorToStore(cmt.GetAuthor())
		res.Committer = g.commitAuthorToStore(cmt.GetCommitter())
		res.ParentSHAs = transform(cmt.Parents, func(c *gh.Commit) string { return c.GetSHA() })
	}
	return res
}

type ghUser interface {
	GetLogin() string
	GetEmail() string
}

type dateGetter interface {
	GetDate() time.Time
}

func (g *Github) commitAuthorToStore(user ghUser) git.User {
	res := git.User{
		Username: user.GetLogin(),
		Email:    user.GetEmail(),
	}

	if dg, ok := user.(dateGetter); ok {
		res.Date = dg.GetDate()
	}

	return res
}
