package engine

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Semior001/releaseit/app/git"
	"github.com/go-pkgz/requester"
	"github.com/samber/lo"
	gl "github.com/xanzy/go-gitlab"
)

// Gitlab implements Interface with gitlab API below it.
type Gitlab struct {
	cl        *gl.Client
	projectID string
}

// NewGitlab creates a new Gitlab engine.
func NewGitlab(token, baseURL, projectID string, httpCl http.Client) (*Gitlab, error) {
	var (
		cl  = requester.New(httpCl)
		svc = &Gitlab{
			projectID: projectID,
		}
		err error
	)

	svc.cl, err = gl.NewClient(
		token,
		gl.WithBaseURL(baseURL),
		gl.WithHTTPClient(cl.Client()),
	)
	if err != nil {
		return nil, fmt.Errorf("initialize gitlab client: %w", err)
	}

	return svc, nil
}

// Compare two commits by their SHA.
func (g *Gitlab) Compare(ctx context.Context, fromSHA, toSHA string) (git.CommitsComparison, error) {
	opts := &gl.CompareOptions{From: &fromSHA, To: &toSHA}

	cmp, _, err := g.cl.Repositories.Compare(g.projectID, opts, gl.WithContext(ctx))
	if err != nil {
		return git.CommitsComparison{}, fmt.Errorf("do request: %w", err)
	}

	commits := make([]git.Commit, len(cmp.Commits))
	for i, commit := range cmp.Commits {
		commits[i] = g.commitToStore(commit)
	}

	return git.CommitsComparison{
		Commits:      commits,
		TotalCommits: len(commits),
	}, nil
}

// ListPRsOfCommit returns pull requests associated with commit by the given SHA.
func (g *Gitlab) ListPRsOfCommit(ctx context.Context, sha string) ([]git.PullRequest, error) {
	mrs, _, err := g.cl.Commits.ListMergeRequestsByCommit(g.projectID, sha, gl.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	res := make([]git.PullRequest, len(mrs))
	for i, mr := range mrs {
		res[i] = git.PullRequest{
			Number:   mr.IID,
			Title:    mr.Title,
			Body:     mr.Description,
			Author:   git.User{Username: lo.FromPtr(mr.Author).Username},
			Labels:   mr.Labels,
			ClosedAt: lo.FromPtr(mr.MergedAt),
		}
	}

	return res, nil
}

// ListTags returns all tags of the repository.
func (g *Gitlab) ListTags(ctx context.Context) ([]git.Tag, error) {
	opts := &gl.ListTagsOptions{OrderBy: gl.String("updated"), Sort: gl.String("desc")}
	tags, _, err := g.cl.Tags.ListTags(g.projectID, opts, gl.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("github returned error: %w", err)
	}

	res := make([]git.Tag, len(tags))

	for i, tag := range tags {
		res[i] = git.Tag{
			Name:   tag.Name,
			Commit: g.commitToStore(tag.Commit),
		}
	}

	return res, nil
}

// HeadCommit returns the alias of the oldest commit in the repository
func (g *Gitlab) HeadCommit(_ context.Context) (string, error) {
	return "HEAD", nil
}

func (g *Gitlab) commitToStore(commit *gl.Commit) git.Commit {
	return git.Commit{
		SHA: commit.ID,
		Committer: git.User{
			Date:     lo.FromPtr(commit.CommittedDate),
			Username: commit.CommitterName,
			Email:    commit.CommitterEmail,
		},
		Author: git.User{
			Date:     lo.FromPtr(commit.AuthoredDate),
			Username: commit.AuthorName,
			Email:    commit.AuthorEmail,
		},
		ParentSHAs: commit.ParentIDs,
	}
}