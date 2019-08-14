package client

import (
	"context"
	"errors"
	"strings"
)

// GetPullRequestLatestCommitHash gets latest commit hash in PR
func (c *Client) GetPullRequestLatestCommitHash(
	ctx context.Context,
	owner string,
	repo string,
	number int,
) (string, error) {
	commits, _, err := c.client.PullRequests.ListCommits(ctx, owner, repo, number, nil)
	if err != nil {
		return "", err
	}
	if len(commits) == 0 {
		return "", errors.New("PR has no commits")
	}
	return commits[len(commits)-1].GetSHA(), nil
}

// GetPullRequestBaseBranch gets base branch of pull request
func (c *Client) GetPullRequestBaseBranch(
	ctx context.Context,
	owner string,
	repo string,
	number int,
) (string, error) {
	pr, _, err := c.client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return "", err
	}
	// Trim username from username:branch
	parts := strings.Split(pr.GetBase().GetLabel(), ":")
	return parts[len(parts)-1], nil
}

func (c *Client) GetOpenPullRequestNumbers(
	ctx context.Context,
	owner string,
	repo string,
) ([]int, error) {
	prs, _, err := c.client.PullRequests.List(ctx, owner, repo, nil)
	if err != nil {
		return nil, err
	}
	numbers := make([]int, 0, len(prs))
	for _, pr := range prs {
		numbers = append(numbers, pr.GetNumber())
	}
	return numbers, nil
}
