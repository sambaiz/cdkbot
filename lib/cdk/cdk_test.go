package cdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientSetup(t *testing.T) {
	err := new(Client).Setup("./test_repository")
	assert.Nil(t, err)
}

func TestClientList(t *testing.T) {
	lists, err := new(Client).List("./test_repository", map[string]string{"env": "stg"})
	assert.Nil(t, err)
	assert.Equal(t, []string{"Stack1", "Stack2"}, lists)
}

func TestClientDiff(t *testing.T) {
	result, hasDiff := new(Client).Diff("./test_repository", "stack1 stack2", map[string]string{"env": "stg", "foo": "bar"})
	assert.True(t, hasDiff)
	assert.Equal(t, "diff: diff stack1 stack2 -c env=stg foo=bar", result)
}

func TestClientDeploy(t *testing.T) {
	result, err := new(Client).Deploy("./test_repository", "stack1 stack2", map[string]string{"env": "stg", "foo": "bar"})
	assert.Nil(t, err)
	assert.Equal(t, "deploy: deploy --require-approval never stack1 stack2 -c env=stg foo=bar", result)
}
