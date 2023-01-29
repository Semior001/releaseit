// Package engine contains interfaces for different git providers.
package engine

import (
	"context"

	"github.com/Semior001/releaseit/app/git"
)

//go:generate moq -out mock_interface.go . Interface

// Interface defines methods to retrieve information about repository.
type Interface interface {
	// Compare returns comparison between two commits,
	// given by their SHA.
	Compare(ctx context.Context, fromSHA, toSHA string) (git.CommitsComparison, error)
	// ListPRsOfCommit returns pull/merge requests
	// associated with the commit, given by its SHA.
	ListPRsOfCommit(ctx context.Context, sha string) ([]git.PullRequest, error)
	// ListTags returns tags of the repository in descending order of creation.
	ListTags(ctx context.Context) ([]git.Tag, error)
	// GetLastCommitOfBranch returns the SHA or alias of the last commit in the branch.
	GetLastCommitOfBranch(ctx context.Context, branch string) (string, error)
}
