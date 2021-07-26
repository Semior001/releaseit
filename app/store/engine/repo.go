// Package engine defines interfaces each supported repository provider should implement.
// Includes default implementation with github.
package engine

import (
	"context"

	"github.com/Semior001/releaseit/app/store"
)

// Interface defines methods to retrieve information about repository.
type Interface interface {
	// Compare returns comparison between two commits,
	// given by their SHA.
	Compare(ctx context.Context, fromSHA, toSHA string) (store.CommitsComparison, error)
	// ListPRsOfCommit returns pull/merge requests
	// associated with the commit, given by its SHA.
	ListPRsOfCommit(ctx context.Context, sha string) ([]store.PullRequest, error)
	// ListTags returns tags of the repository in descending order of creation.
	ListTags(ctx context.Context) ([]store.Tag, error)
	// HeadCommit returns the SHA or alias of the oldest commit in the repository
	HeadCommit(ctx context.Context) (string, error)
}
