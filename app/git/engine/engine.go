// Package engine contains interfaces for different git providers.
package engine

import (
	"context"
	"errors"
	"github.com/Semior001/releaseit/app/git"
	"time"
)

const defaultPingTimeout = 1 * time.Minute

//go:generate rm -f mock_interface.go
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

// Unsupported is a git engine implementation that returns an error for each method.
type Unsupported struct{}

// Compare returns an error.
func (Unsupported) Compare(context.Context, string, string) (git.CommitsComparison, error) {
	return git.CommitsComparison{}, errors.New("operation not supported")
}

// ListPRsOfCommit returns an error.
func (Unsupported) ListPRsOfCommit(context.Context, string) ([]git.PullRequest, error) {
	return nil, errors.New("operation not supported")
}

// ListTags returns an error.
func (Unsupported) ListTags(context.Context) ([]git.Tag, error) {
	return nil, errors.New("operation not supported")
}

// GetLastCommitOfBranch returns an error.
func (Unsupported) GetLastCommitOfBranch(context.Context, string) (string, error) {
	return "", errors.New("operation not supported")
}
