package engine

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnsupported_Compare(t *testing.T) {
	res, err := Unsupported{}.Compare(nil, "", "")
	assert.EqualError(t, err, "operation not supported")
	assert.Empty(t, res)
}

func TestUnsupported_ListPRsOfCommit(t *testing.T) {
	res, err := Unsupported{}.ListPRsOfCommit(nil, "")
	assert.EqualError(t, err, "operation not supported")
	assert.Empty(t, res)
}

func TestUnsupported_ListTags(t *testing.T) {
	res, err := Unsupported{}.ListTags(nil)
	assert.EqualError(t, err, "operation not supported")
	assert.Empty(t, res)
}

func TestUnsupported_GetLastCommitOfBranch(t *testing.T) {
	res, err := Unsupported{}.GetLastCommitOfBranch(nil, "")
	assert.EqualError(t, err, "operation not supported")
	assert.Empty(t, res)
}
