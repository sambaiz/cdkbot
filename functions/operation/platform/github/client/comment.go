package client

import (
	"context"

	"github.com/google/go-github/v26/github"
)

// CreateComment creates a comment
func (c *Client) CreateComment(
	ctx context.Context,
	body string,
) error {
	_, _, err := c.client.Issues.CreateComment(ctx, c.owner, c.repo, c.number, &github.IssueComment{
		Body: &body,
	})
	return err
}
