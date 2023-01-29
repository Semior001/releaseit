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
	Number   int       `yaml:"number"`
	Title    string    `yaml:"title"`
	Body     string    `yaml:"body"`
	Author   User      `yaml:"author"`
	Labels   []string  `yaml:"labels"`
	ClosedAt time.Time `yaml:"closed_at"`
	Branch   string    `yaml:"branch"`
	URL      string    `yaml:"url"`
}

// User holds user data.
type User struct {
	Username string `yaml:"username"`
	Email    string `yaml:"email"`
}

// Commit represents a repository commit.
type Commit struct {
	SHA        string
	ParentSHAs []string
	Message    string
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
