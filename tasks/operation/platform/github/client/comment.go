package client

import (
	"context"
	"errors"
	"github.com/google/go-github/v26/github"
	"github.com/sambaiz/cdkbot/tasks/operation/platform"
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

// ListComments gets comments order by posted asc
func (c *Client) ListComments(
	ctx context.Context,
) ([]platform.Comment, error) {
	page := 1
	comments := []*github.IssueComment{}
	for true {
		// Return ID asc. Option's order doesn't seem to work
		paging, _, err := c.client.Issues.ListComments(ctx, c.owner, c.repo, c.number, &github.IssueListCommentsOptions{
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
		comments = append(comments, paging...)
		page++
		if page > maxPage {
			return nil, errors.New("Too many comments")
		}
	}
	ret := []platform.Comment{}
	for _, comment := range comments {
		ret = append(ret, platform.Comment{
			ID:   comment.GetID(),
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
	_, err := c.client.Issues.DeleteComment(ctx, c.owner, c.repo, commentID)
	if err != nil {
		return err
	}
	return nil
}
