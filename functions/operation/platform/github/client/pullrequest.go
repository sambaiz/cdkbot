package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/v26/github"
	"github.com/sambaiz/cdkbot/functions/operation/constant"
	"strings"
)

// GetPullRequestLatestCommitHash gets latest commit hash of PR
func (c *Client) GetPullRequestLatestCommitHash(ctx context.Context) (string, error) {
	page := 1
	commits := []*github.RepositoryCommit{}
	for true {
		paging, _, err := c.client.PullRequests.ListCommits(ctx, c.owner, c.repo, c.number, &github.ListOptions{
			Page:    page,
			PerPage: 100,
		})
		if err != nil {
			return "", err
		}
		if len(paging) == 0 {
			break
		}
		commits = append(commits, paging...)
		page++
		// API can't return more than 250 commits for a pull request
		if len(commits) >= 250 {
			return "", errors.New("Too many commits")
		}
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

// MergePullRequest merges PR
func (c *Client) MergePullRequest(ctx context.Context, message string) error {
	_, _, err := c.client.PullRequests.Merge(ctx, c.owner, c.repo, c.number, message, nil)
	return err
}

func (c *Client) getOpenPullRequestNumbers(
	ctx context.Context,
) ([]int, error) {
	page := 1
	prs := []*github.PullRequest{}
	for true {
		paging, _, err := c.client.PullRequests.List(ctx, c.owner, c.repo, &github.PullRequestListOptions{
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: 100,
			},
		})
		if err != nil {
			return nil, err
		}
		if len(paging) == 0 {
			break
		}
		prs = append(prs, paging...)
		page++
		if page > maxPage {
			return nil, fmt.Errorf("Too many PRs")
		}
	}
	numbers := make([]int, 0, len(prs))
	for _, pr := range prs {
		numbers = append(numbers, pr.GetNumber())
	}
	return numbers, nil
}
