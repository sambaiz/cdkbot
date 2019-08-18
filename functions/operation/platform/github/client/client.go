package client

import (
	"context"
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
