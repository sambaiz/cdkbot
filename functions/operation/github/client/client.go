package client

import (
	"context"
	"os"

	"github.com/google/go-github/v26/github"
	"golang.org/x/oauth2"
)

// Clienter is interface of GitHub Client
type Clienter interface {
	CreateStatusOfLatestCommit(
		ctx context.Context,
		owner string,
		repo string,
		number int,
		state State,
		description *string,
	) error
	CreateComment(
		ctx context.Context,
		owner string,
		repo string,
		number int,
		body string,
	) error
	GetPullRequestLatestCommitHash(
		ctx context.Context,
		owner string,
		repo string,
		number int,
	) (string, error)
	GetPullRequestBaseBranch(
		ctx context.Context,
		owner string,
		repo string,
		number int,
	) (string, error)
}

// Client is struct of GitHub Client
type Client struct {
	client *github.Client
}

// New GitHub client
func New(ctx context.Context) *Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &Client{github.NewClient(tc)}
}
