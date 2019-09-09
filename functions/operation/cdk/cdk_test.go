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
	type expected struct {
		outResult  string
		outHasDiff bool
		isError    bool
	}
	tests := []struct {
		title    string
		inStacks string
		expected expected
	}{
		{
			title:    "has no diff",
			inStacks: "stack1 stack2",
			expected: expected{
				outResult:  "diff: diff stack1 stack2 -c env=stg",
				outHasDiff: false,
				isError:    false,
			},
		},
		{
			title:    "has diff",
			inStacks: "diffStack",
			expected: expected{
				outResult:  "diff: diff diffStack -c env=stg",
				outHasDiff: true,
				isError:    false,
			},
		},
		{
			title:    "error",
			inStacks: "failStack",
			expected: expected{
				outResult:  "failed!",
				outHasDiff: true,
				isError:    true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			result, hasDiff, err := new(Client).Diff("./test_repository", test.inStacks, map[string]string{"env": "stg"})
			assert.Equal(t, test.expected.outResult, result)
			assert.Equal(t, hasDiff, test.expected.outHasDiff)
			assert.Equal(t, test.expected.isError, err != nil)
		})
	}
}

func TestClientDeploy(t *testing.T) {
	type expected struct {
		outResult string
		isError   bool
	}
	tests := []struct {
		title    string
		inStacks string
		expected expected
	}{
		{
			title:    "success",
			inStacks: "stack1 stack2",
			expected: expected{
				outResult: "deploy: deploy stack1 stack2 --require-approval never -c env=stg",
				isError:   false,
			},
		},
		{
			title:    "error",
			inStacks: "failStack",
			expected: expected{
				outResult: "failed!",
				isError:   true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			result, err := new(Client).Deploy("./test_repository", test.inStacks, map[string]string{"env": "stg"})
			assert.Equal(t, test.expected.outResult, result)
			assert.Equal(t, test.expected.isError, err != nil)
		})
	}
}
