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

// Empty returns true if user is empty.
func (u User) Empty() bool {
	return u.Date.IsZero() && u.Username == "" && u.Email == ""
}

// Commit represents a repository commit.
type Commit struct {
	SHA        string
	Committer  User
	Author     User
	ParentSHAs []string
}

// Empty returns true if SHA of the commit is not specified.
func (c Commit) Empty() bool {
	return c.SHA == ""
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

// Empty returns true if the tag is empty.
func (t Tag) Empty() bool {
	return t.Name == "" && t.Commit.Empty()
}
