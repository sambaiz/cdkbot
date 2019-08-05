package cdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetup(t *testing.T) {
	err := new(Client).Setup("./test")
	assert.Nil(t, err)
}

func TestList(t *testing.T) {
	lists, err := new(Client).List("./test")
	assert.Nil(t, err)
	assert.Equal(t, []string{"Stack1", "Stack2"}, lists)
}

func TestDiff(t *testing.T) {
	result, hasDiff := new(Client).Diff("./test")
	assert.True(t, hasDiff)
	assert.Equal(t, "diff\ndiff", result)
}

func TestDeploy(t *testing.T) {
	result, err := new(Client).Deploy("./test", "stack")
	assert.Nil(t, err)
	assert.Equal(t, "deploy\ndeploy", result)
}
