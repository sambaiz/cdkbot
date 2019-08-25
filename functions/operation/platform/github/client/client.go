package client

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v26/github"
	"golang.org/x/oauth2"
)

// Client is an implementation of platform.Client
type Client struct {
	client *github.Client
	owner  string
	repo   string
	number int
}

// New GitHub client
func New(
	ctx context.Context,
	owner string,
	repo string,
	number int,
) *Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &Client{
		client: github.NewClient(tc),
		owner:  owner,
		repo:   repo,
		number: number,
	}
}

// NewWithHeadBranch create GitHub client with head branch
func NewWithHeadBranch(
	ctx context.Context,
	owner string,
	repo string,
	headBranch string,
) (*Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	prs, _, err := client.PullRequests.List(ctx, owner, repo, &github.PullRequestListOptions{
		Head:        headBranch,
	})
	if err != nil {
		return nil, err
	}
	if len(prs) == 0 {
		return nil, fmt.Errorf("PR is not found with head branch %s", headBranch)
	}
	return New(ctx, owner, repo, prs[0].GetNumber()), nil
}

const maxPage  = 50