package engine

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Semior001/releaseit/app/store"
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

// BasicAuth describes basic authentication parameters
// for API requests.
type BasicAuth struct {
	Username string
	Password string
}

// Empty returns true if BasicAuth is empty.
func (b BasicAuth) Empty() bool { return b.Username == "" && b.Password == "" }

// NewGithub makes new instance of Github.
func NewGithub(owner, name string, httpCl http.Client, basicAuth BasicAuth) *Github {
	svc := &Github{
		owner: owner,
		name:  name,
	}

	cl := requester.New(httpCl)

	if !basicAuth.Empty() {
		cl.Use(middleware.BasicAuth(basicAuth.Username, basicAuth.Password))
	}

	svc.cl = gh.NewClient(cl.Client())

	return svc
}

// Compare two commits by their SHA.
func (g *Github) Compare(ctx context.Context, fromSHA, toSHA string) (store.CommitsComparison, error) {
	comp, _, err := g.cl.Repositories.CompareCommits(ctx, g.owner, g.name, fromSHA, toSHA)
	if err != nil {
		return store.CommitsComparison{}, fmt.Errorf("github returned error: %w", err)
	}

	commits := make([]store.Commit, len(comp.Commits))

	for i, commit := range comp.Commits {
		commits[i] = g.commitToStore(commit)
	}

	return store.CommitsComparison{
		Commits:      commits,
		TotalCommits: comp.GetTotalCommits(),
	}, nil
}

// ListPRsOfCommit returns pull requests associated with commit by the given SHA.
func (g *Github) ListPRsOfCommit(ctx context.Context, sha string) ([]store.PullRequest, error) {
	prs, _, err := g.cl.PullRequests.ListPullRequestsWithCommit(ctx, g.owner, g.name, sha, &gh.PullRequestListOptions{})
	if err != nil {
		return nil, fmt.Errorf("github returned error: %w", err)
	}

	res := make([]store.PullRequest, len(prs))

	for i, pr := range prs {
		res[i] = store.PullRequest{
			Number:   pr.GetNumber(),
			Title:    pr.GetTitle(),
			Body:     pr.GetBody(),
			ClosedAt: pr.GetClosedAt(),
			Author:   store.User{Username: pr.GetUser().GetLogin(), Email: pr.GetUser().GetEmail()},
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
func (g *Github) ListTags(ctx context.Context) ([]store.Tag, error) {
	tags, _, err := g.cl.Repositories.ListTags(ctx, g.owner, g.name, &gh.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("github returned error: %w", err)
	}

	res := make([]store.Tag, len(tags))

	for i, tag := range tags {
		res[i] = store.Tag{
			Name:   tag.GetName(),
			Commit: g.commitToStore(tag.GetCommit()),
		}
	}

	return res, nil
}

type shaGetter interface {
	GetSHA() string
}

func (g *Github) commitToStore(commitInterface shaGetter) store.Commit {
	res := store.Commit{SHA: commitInterface.GetSHA()}
	switch commit := commitInterface.(type) {
	case *gh.Commit:
		res.Author = g.commitAuthorToStore(commit.GetAuthor())
		res.Committer = g.commitAuthorToStore(commit.GetCommitter())

		res.Parents = make([]store.Commit, len(commit.Parents))
		for i, parent := range commit.Parents {
			res.Parents[i] = g.commitToStore(parent)
		}
	case *gh.RepositoryCommit:
		res.Author = g.commitAuthorToStore(commit.GetAuthor())
		res.Committer = g.commitAuthorToStore(commit.GetCommitter())

		res.Parents = make([]store.Commit, len(commit.Parents))
		for i, parent := range commit.Parents {
			res.Parents[i] = g.commitToStore(parent)
		}
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

func (g *Github) commitAuthorToStore(user ghUser) store.User {
	res := store.User{
		Username: user.GetLogin(),
		Email:    user.GetEmail(),
	}

	if dg, ok := user.(dateGetter); ok {
		res.Date = dg.GetDate()
	}

	return res
}
