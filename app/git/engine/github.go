package engine

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Semior001/releaseit/app/git"
	"github.com/go-pkgz/requester"
	"github.com/go-pkgz/requester/middleware"
	"github.com/go-pkgz/requester/middleware/logger"
	gh "github.com/google/go-github/v37/github"
	"github.com/samber/lo"
)

// Github implements Repository with github API below it.
type Github struct {
	cl    *gh.Client
	owner string
	name  string
}

// GithubParams contains parameters for github engine.
type GithubParams struct {
	Owner             string
	Name              string
	BasicAuthUsername string
	BasicAuthPassword string
	HTTPClient        http.Client
}

// NewGithub makes new instance of Github.
func NewGithub(ctx context.Context, params GithubParams) (*Github, error) {
	svc := &Github{
		owner: params.Owner,
		name:  params.Name,
	}

	cl := requester.New(params.HTTPClient, logger.New(logger.Func(log.Printf), logger.Prefix("[DEBUG]")).Middleware)

	if params.BasicAuthUsername != "" && params.BasicAuthPassword != "" {
		cl.Use(middleware.BasicAuth(params.BasicAuthUsername, params.BasicAuthPassword))
	}

	svc.cl = gh.NewClient(cl.Client())

	ctx, cancel := context.WithTimeout(ctx, defaultPingTimeout)
	defer cancel()

	if _, _, err := svc.cl.Repositories.Get(ctx, svc.owner, svc.name); err != nil {
		return nil, fmt.Errorf("check connection to github: %w", err)
	}

	return svc, nil
}

// GetLastCommitOfBranch returns the SHA or alias of the last commit in the branch.
func (g *Github) GetLastCommitOfBranch(ctx context.Context, branchName string) (string, error) {
	branch, _, err := g.cl.Repositories.GetBranch(ctx, g.owner, g.name, branchName, true)
	if err != nil {
		return "", fmt.Errorf("get branch: %w", err)
	}

	return branch.GetCommit().GetSHA(), nil
}

// Compare two commits by their SHA.
func (g *Github) Compare(ctx context.Context, fromSHA, toSHA string) (git.CommitsComparison, error) {
	comp, _, err := g.cl.Repositories.CompareCommits(ctx, g.owner, g.name, fromSHA, toSHA)
	if err != nil {
		return git.CommitsComparison{}, fmt.Errorf("github returned error: %w", err)
	}

	commits := make([]git.Commit, len(comp.Commits))

	for i, commit := range comp.Commits {
		commits[i] = g.transformCommit(commit)
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
		return nil, fmt.Errorf("list pull requests with commit: %w", err)
	}

	res := make([]git.PullRequest, len(prs))

	for i, pr := range prs {
		res[i] = g.transformPR(pr)

		for _, assignee := range pr.Assignees {
			res[i].Assignees = append(res[i].Assignees, git.User{
				Username: assignee.GetLogin(),
				Email:    assignee.GetEmail(),
			})
		}
	}

	return res, nil
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
			Commit: g.transformCommit(tag.GetCommit()),
		}
	}

	return res, nil
}

type shaGetter interface {
	GetSHA() string
}

func (g *Github) transformCommit(commitInterface shaGetter) git.Commit {
	res := git.Commit{SHA: commitInterface.GetSHA()}
	switch cmt := commitInterface.(type) {
	case *gh.Commit:
		res.ParentSHAs = lo.Map(cmt.Parents, func(c *gh.Commit, _ int) string { return c.GetSHA() })
		res.Message = cmt.GetMessage()
		res.CommittedAt = cmt.GetCommitter().GetDate()
		res.AuthoredAt = cmt.GetAuthor().GetDate()
		res.URL = cmt.GetURL()
		res.Author = git.User{Username: cmt.GetAuthor().GetLogin(), Email: cmt.GetAuthor().GetEmail()}
		res.Committer = git.User{Username: cmt.GetCommitter().GetLogin(), Email: cmt.GetCommitter().GetEmail()}
	case *gh.RepositoryCommit:
		res.ParentSHAs = lo.Map(cmt.Parents, func(c *gh.Commit, _ int) string { return c.GetSHA() })
		res.Message = cmt.GetCommit().GetMessage()
		res.CommittedAt = cmt.GetCommit().GetCommitter().GetDate()
		res.AuthoredAt = cmt.GetCommit().GetAuthor().GetDate()
		res.URL = cmt.GetURL()
		res.Author = git.User{Username: cmt.GetAuthor().GetLogin(), Email: cmt.GetAuthor().GetEmail()}
		res.Committer = git.User{Username: cmt.GetCommitter().GetLogin(), Email: cmt.GetCommitter().GetEmail()}
	}
	return res
}

func (g *Github) transformPR(pr *gh.PullRequest) git.PullRequest {
	return git.PullRequest{
		Number:       pr.GetNumber(),
		Title:        pr.GetTitle(),
		Body:         pr.GetBody(),
		ClosedAt:     pr.GetClosedAt(),
		Author:       git.User{Username: pr.GetUser().GetLogin(), Email: pr.GetUser().GetEmail()},
		Labels:       lo.Map(pr.Labels, func(l *gh.Label, _ int) string { return l.GetName() }),
		SourceBranch: pr.GetHead().GetRef(),
		TargetBranch: pr.GetBase().GetRef(),
		URL:          pr.GetHTMLURL(),
	}
}
