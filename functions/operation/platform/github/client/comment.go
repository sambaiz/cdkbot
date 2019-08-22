package client

import (
	"context"
	"github.com/google/go-github/v26/github"
	"github.com/sambaiz/cdkbot/functions/operation/platform"
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

// ListComments gets comments in order of creation
func (c *Client) ListComments(
	ctx context.Context,
) ([]platform.Comment, error) {
	comments, _, err := c.client.PullRequests.ListComments(ctx, c.owner, c.repo, c.number, &github.PullRequestListCommentsOptions{
		Sort:        "created",
		Direction:   "asc",
	})
	if err != nil {
		return nil, err
	}
	ret := []platform.Comment{}
	for _, comment := range comments {
		ret = append(ret, platform.Comment{
			ID: comment.GetID(),
			Body: comment.GetBody(),
		})
	}
	return ret, nil
}

// DeleteComment deletes a comment
func (c *Client) DeleteComment(
	ctx context.Context,
	commentID int64,
) error {
	_, err := c.client.PullRequests.DeleteComment(ctx, c.owner, c.repo, commentID)
	if err != nil {
		return err
	}
	return nil
}
