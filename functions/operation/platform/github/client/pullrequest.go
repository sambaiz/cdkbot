package client

import (
	"context"
	"errors"
	"github.com/sambaiz/cdkbot/functions/operation/constant"
	"strings"
)

// GetPullRequestLatestCommitHash gets latest commit hash of PR
func (c *Client) GetPullRequestLatestCommitHash(ctx context.Context) (string, error) {
	commits, _, err := c.client.PullRequests.ListCommits(ctx, c.owner, c.repo, c.number, nil)
	if err != nil {
		return "", err
	}
	if len(commits) == 0 {
		return "", errors.New("PR has no commits")
	}
	return commits[len(commits)-1].GetSHA(), nil
}

// GetPullRequestBaseBranch gets base branch of PR
func (c *Client) GetPullRequestBaseBranch(
	ctx context.Context,
) (string, error) {
	pr, _, err := c.client.PullRequests.Get(ctx, c.owner, c.repo, c.number)
	if err != nil {
		return "", err
	}
	// Trim username from username:branch
	parts := strings.Split(pr.GetBase().GetLabel(), ":")
	return parts[len(parts)-1], nil
}


// GetPullRequestLabels gets PR's labels and returns map[label name]constant.Label
func (c *Client) GetPullRequestLabels(ctx context.Context) (map[string]constant.Label, error) {
	pr, _, err := c.client.PullRequests.Get(ctx, c.owner, c.repo, c.number)
	if err != nil {
		return nil, err
	}
	labels := map[string]constant.Label{}
	for _, label := range pr.Labels {
		if lb, ok := constant.NameToLabel[label.GetName()]; ok {
			labels[lb.Name] = constant.NameToLabel[lb.Name]
		}
	}
	return labels, nil
}

func (c *Client) getOpenPullRequestNumbers(
	ctx context.Context,
) ([]int, error) {
	prs, _, err := c.client.PullRequests.List(ctx, c.owner, c.repo, nil)
	if err != nil {
		return nil, err
	}
	numbers := make([]int, 0, len(prs))
	for _, pr := range prs {
		numbers = append(numbers, pr.GetNumber())
	}
	return numbers, nil
}
