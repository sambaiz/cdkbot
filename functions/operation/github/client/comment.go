package client

import (
	"context"

	"github.com/google/go-github/v26/github"
)

// CreateComment creates comment
func (c *Client) CreateComment(
	ctx context.Context,
	owner string,
	repo string,
	number int,
	body string,
) error {
	_, _, err := c.client.Issues.CreateComment(ctx, owner, repo, number, &github.IssueComment{
		Body: &body,
	})
	return err
}
