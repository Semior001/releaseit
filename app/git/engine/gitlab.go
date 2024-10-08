package engine

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Semior001/releaseit/app/git"
	"github.com/go-pkgz/requester"
	"github.com/go-pkgz/requester/middleware/logger"
	"github.com/samber/lo"
	gl "github.com/xanzy/go-gitlab"
)

// Gitlab implements Repository with gitlab API below it.
type Gitlab struct {
	cl        *gl.Client
	projectID string
}

// NewGitlab creates a new Gitlab engine.
func NewGitlab(ctx context.Context, token, baseURL, projectID string, httpCl http.Client) (*Gitlab, error) {
	var (
		cl  = requester.New(httpCl, logger.New(logger.Func(log.Printf), logger.Prefix("[DEBUG]")).Middleware)
		svc = &Gitlab{projectID: projectID}
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

	ctx, cancel := context.WithTimeout(ctx, defaultPingTimeout)
	defer cancel()

	if _, _, err = svc.cl.Projects.GetProject(projectID, &gl.GetProjectOptions{}, gl.WithContext(ctx)); err != nil {
		return nil, fmt.Errorf("ping gitlab: %w", err)
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
		commits[i] = g.transformCommit(commit)
	}

	return git.CommitsComparison{
		Commits:      commits,
		TotalCommits: len(commits),
	}, nil
}

// GetLastCommitOfBranch returns the SHA or alias of the last commit in the branch.
func (g *Gitlab) GetLastCommitOfBranch(ctx context.Context, branchName string) (string, error) {
	branch, _, err := g.cl.Branches.GetBranch(g.projectID, branchName, gl.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}

	return branch.Commit.ID, nil
}

// ListPRsOfCommit returns pull requests associated with commit by the given SHA.
func (g *Gitlab) ListPRsOfCommit(ctx context.Context, sha string) ([]git.PullRequest, error) {
	mrs, _, err := g.cl.Commits.ListMergeRequestsByCommit(g.projectID, sha, gl.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	res := make([]git.PullRequest, len(mrs))
	for i, mr := range mrs {
		res[i] = g.transformMR(mr)

		for _, assignee := range mr.Assignees {
			res[i].Assignees = append(res[i].Assignees, git.User{Username: assignee.Username})
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
			Commit: g.transformCommit(tag.Commit),
		}
	}

	return res, nil
}

func (g *Gitlab) transformCommit(commit *gl.Commit) git.Commit {
	return git.Commit{
		SHA:         commit.ID,
		ParentSHAs:  commit.ParentIDs,
		Message:     commit.Message,
		CommittedAt: lo.FromPtr(commit.CommittedDate),
		AuthoredAt:  lo.FromPtr(commit.AuthoredDate),
		URL:         commit.WebURL,
		Author: git.User{
			Username: commit.AuthorName,
			Email:    commit.AuthorEmail,
		},
		Committer: git.User{
			Username: commit.CommitterName,
			Email:    commit.CommitterEmail,
		},
	}
}

func (g *Gitlab) transformMR(mr *gl.MergeRequest) git.PullRequest {
	return git.PullRequest{
		Number: mr.IID,
		Title:  mr.Title,
		Body:   mr.Description,
		Author: git.User{Username: lo.FromPtr(mr.Author).Username},
		// FIXME: by some reason, library encodes labels as a string, not a slice.
		Labels: lo.Flatten(lo.Map(mr.Labels, func(s string, _ int) []string {
			return strings.Split(s, ",")
		})),
		// closed at in MR points to time when MR was closed without merging,
		// so we use merged at instead.
		ClosedAt:     lo.FromPtr(mr.MergedAt),
		SourceBranch: mr.SourceBranch,
		TargetBranch: mr.TargetBranch,
		URL:          mr.WebURL,
	}
}
