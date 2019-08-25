package client

import (
	"context"
	"github.com/google/go-github/v26/github"
	"github.com/sambaiz/cdkbot/functions/operation/constant"
)

var stateMap = map[constant.State]*string{
	constant.StateMergeReady: &[]string{"success"}[0],
	constant.StateNeedDeploy: &[]string{"failure"}[0],
	constant.StateRunning:    &[]string{"pending"}[0],
	constant.StateError:      &[]string{"error"}[0],
}

var statusContext = "cdkbot"

// SetStatus set status of latest commit
func (c *Client) SetStatus(
	ctx context.Context,
	state constant.State,
	description string,
) error {
	pr, err := c.GetPullRequest(ctx)
	if err != nil {
		return err
	}

	_, _, err = c.client.Repositories.CreateStatus(ctx, c.owner, c.repo, pr.HeadCommitHash, &github.RepoStatus{
		State:       stateMap[state],
		Context:     &statusContext,
		Description: &description,
	})
	if err != nil {
		return err
	}
	return nil
}
