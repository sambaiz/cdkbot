package client

import (
	"context"
	"github.com/google/go-github/v26/github"
	"github.com/sambaiz/cdkbot/functions/operation/constant"
)

// AddLabel adds label to PR
func (c *Client) AddLabel(
	ctx context.Context,
	label constant.Label,
) error {
	return c.addLabel(ctx, c.number, label)
}

// AddLabelToOtherPRs adds label to other PRs
func (c *Client) AddLabelToOtherPRs(
	ctx context.Context,
	label constant.Label,
) error {
	numbers, err := c.getOpenPullRequestNumbers(ctx)
	if err != nil {
		return err
	}
	for _, number := range numbers {
		if c.number == number {
			continue
		}
		if err := c.addLabel(ctx, number, label); err != nil {
			return err
		}
	}
	return nil
}

// RemoveLabel removes label from PR
func (c *Client) RemoveLabel(
	ctx context.Context,
	label constant.Label,
) error {
	// ignore error "does not exist"
	c.client.Issues.RemoveLabelForIssue(ctx, c.owner, c.repo, c.number, label.Name)
	return nil
}

func (c *Client) addLabel(ctx context.Context, number int, label constant.Label) error {
	added, _, err := c.client.Issues.AddLabelsToIssue(ctx, c.owner, c.repo, c.number, []string{label.Name})
	if len(added) == 1 && added[0].GetDescription() == "" {
		label, ok := constant.NameToLabel[added[0].GetName()]
		if !ok {
			return nil
		}
		_, _, err = c.client.Issues.EditLabel(ctx, c.owner, c.repo, added[0].GetName(), &github.Label{
			Name:        &[]string{added[0].GetName()}[0],
			Description: &label.Description,
			Color:       &label.Color,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
