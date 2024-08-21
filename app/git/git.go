// Package git contains types and engines to work with git repositories.
package git

import (
	"time"
)

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
	SHA         string    `yaml:"sha"`
	ParentSHAs  []string  `yaml:"parent_shas"`
	Message     string    `yaml:"message"`
	CommittedAt time.Time `yaml:"committed_at"`
	AuthoredAt  time.Time `yaml:"authored_at"`
	URL         string    `yaml:"url"`
	Author      User      `yaml:"author"`
	Committer   User      `yaml:"committer"`
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
