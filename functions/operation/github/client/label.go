package client

import (
	"context"

	"github.com/google/go-github/v26/github"
)

type Label struct {
	name        string
	description string
	color       string
}

var (
	LabelOutdatedDiff = Label{
		name:        "cdkbot/outdated diff",
		description: "Diffs are outdated. Run /diff again.",
		color:       "e4e669",
	}
	LabelNoDiff = Label{
		name:        "cdkbot/no diffs",
		description: "No diffs. Let's merge!",
		color:       "008672",
	}
	LabelDeploying = Label{
		name:        "cdkbot/deploying",
		description: "Now deploying",
		color:       "0075ca",
	}
	nameToLabel = map[string]Label{
		LabelOutdatedDiff.name: LabelOutdatedDiff,
		LabelNoDiff.name:       LabelNoDiff,
		LabelDeploying.name:    LabelDeploying,
	}
)

// AddLabels adds labels to PR
func (c *Client) AddLabels(
	ctx context.Context,
	owner string,
	repo string,
	number int,
	labels []Label,
) error {
	names := make([]string, 0, len(labels))
	for _, label := range labels {
		names = append(names, label.name)
	}
	added, _, err := c.client.Issues.AddLabelsToIssue(ctx, owner, repo, number, names)
	if err != nil {
		return err
	}
	for _, a := range added {
		if a.GetDescription() == "" {
			label, ok := nameToLabel[a.GetName()]
			if !ok || a.GetDescription() != "" {
				continue
			}
			_, _, err = c.client.Issues.EditLabel(ctx, owner, repo, a.GetName(), &github.Label{
				Name:        &[]string{a.GetName()}[0],
				Description: &label.description,
				Color:       &label.color,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// RemoveLabel removes label from PR
func (c *Client) RemoveLabel(
	ctx context.Context,
	owner string,
	repo string,
	number int,
	label Label,
) error {
	// ignore error "does not exist"
	c.client.Issues.RemoveLabelForIssue(ctx, owner, repo, number, label.name)
	return nil
}
