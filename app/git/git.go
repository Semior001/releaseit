// Package git contains types and engines to work with git repositories.
package git

import (
	"context"
	"time"
)

//go:generate rm -f mock_repository.go
//go:generate moq -out mock_repository.go . Repository

// Repository defines methods to retrieve information about repository.
type Repository interface {
	// Compare returns comparison between two commits,
	// given by their SHA.
	Compare(ctx context.Context, fromSHA, toSHA string) (CommitsComparison, error)
	// ListPRsOfCommit returns pull/merge requests
	// associated with the commit, given by its SHA.
	ListPRsOfCommit(ctx context.Context, sha string) ([]PullRequest, error)
	// ListTags returns tags of the repository in descending order of creation.
	ListTags(ctx context.Context) ([]Tag, error)
	// GetLastCommitOfBranch returns the SHA or alias of the last commit in the branch.
	GetLastCommitOfBranch(ctx context.Context, branch string) (string, error)
}

// PullRequest represents a pull/merge request from the
// remote repository.
type PullRequest struct {
	Number         int       `yaml:"number"`
	Title          string    `yaml:"title"`
	Body           string    `yaml:"body"`
	Author         User      `yaml:"author"`
	Labels         []string  `yaml:"labels"`
	ClosedAt       time.Time `yaml:"closed_at"`
	SourceBranch   string    `yaml:"source_branch"`
	TargetBranch   string    `yaml:"target_branch"`
	URL            string    `yaml:"url"`
	ReceivedBySHAs []string  `yaml:"received_by_shas"`
	Assignees      []User    `yaml:"assignees"`
}

// User holds user data.
type User struct {
	Username string `yaml:"username"`
	Email    string `yaml:"email"`
}

// Commit represents a repository commit.
type Commit struct {
	SHA         string
	ParentSHAs  []string
	Message     string
	CommittedAt time.Time
	AuthoredAt  time.Time
}

// CommitsComparison is the result of comparing two commits.
type CommitsComparison struct {
	Commits      []Commit
	TotalCommits int
}

// Tag represents a repository tag.
type Tag struct {
	Name   string
	Commit Commit
}
