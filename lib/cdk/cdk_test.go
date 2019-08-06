package cdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	err := new(Client).Setup("./test_repository")
	assert.Nil(t, err)
}

func TestList(t *testing.T) {
	lists, err := new(Client).List("./test_repository")
	assert.Nil(t, err)
	assert.Equal(t, []string{"Stack1", "Stack2"}, lists)
}

func TestClientDiff(t *testing.T) {
	result, hasDiff := new(Client).Diff("./test_repository")
	assert.True(t, hasDiff)
	assert.Equal(t, "diff\ndiff", result)
}

func TestClientDeploy(t *testing.T) {
	result, err := new(Client).Deploy("./test_repository", "stack")
	assert.Nil(t, err)
	assert.Equal(t, "deploy\ndeploy", result)
}
