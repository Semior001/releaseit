// Package git contains types and engines to work with git repositories.
package git

import "time"

// Changelog represents a basic changelog.
type Changelog struct {
	TagName   string
	ClosedPRs []PullRequest
}

// PullRequest represents a pull/merge request from the
// remote repository.
type PullRequest struct {
	Number   int
	Title    string
	Body     string
	Author   User
	Labels   []string
	ClosedAt time.Time
	Branch   string
	URL      string
}

// User holds user data.
type User struct {
	Date     time.Time
	Username string
	Email    string
}

// Commit represents a repository commit.
type Commit struct {
	SHA        string
	Committer  User
	Author     User
	ParentSHAs []string
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
