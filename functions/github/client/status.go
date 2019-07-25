package client

import (
	"context"

	"github.com/google/go-github/v26/github"
)

// State is GitHub's one
type State string

var (
	// StateError means some operations has error
	StateError State = "error"
	// StateFailure means some operations have been failed
	StateFailure State = "failure"
	// StatePending means some operation are running
	StatePending State = "pending"
	// StateSuccess means some operations have been succeeded
	StateSuccess State = "success"
)

var statusContext = "cdkbot"

// CreateStatusOfLatestCommit creates statuses to latest commit
func (c *Client) CreateStatusOfLatestCommit(
	ctx context.Context,
	owner string,
	repo string,
	number int,
	state State,
) error {
	hash, err := c.GetPullRequestLatestCommitHash(ctx, owner, repo, number)
	if err != nil {
		return err
	}

	st := string(state)
	_, _, err = c.client.Repositories.CreateStatus(ctx, owner, repo, hash, &github.RepoStatus{
		State:   &st,
		Context: &statusContext,
	})
	if err != nil {
		return err
	}
	return nil
}
