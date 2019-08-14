package eventhandler

import (
	"context"
	"fmt"

	"github.com/google/go-github/v26/github"
	"github.com/sambaiz/cdkbot/functions/operation/github/client"
)

type pullRequestEvent struct {
	ownerName string
	repoName  string
	prNumber  int
	cloneURL  string
}

// PullRequest handles github.PullRequestEvent
func (e *EventHandler) PullRequest(
	ctx context.Context,
	ev *github.PullRequestEvent,
) error {
	event := pullRequestEvent{
		ownerName: ev.GetRepo().GetOwner().GetLogin(),
		repoName:  ev.GetRepo().GetName(),
		prNumber:  ev.GetPullRequest().GetNumber(),
		cloneURL:  ev.GetRepo().GetCloneURL(),
	}
	var f func() (client.State, string, error)
	switch ev.GetAction() {
	case "opened":
		f = func() (client.State, string, error) {
			return e.pullRequestOpened(ctx, event)
		}
	default:
		return nil
	}
	return e.updateStatus(
		ctx,
		event.ownerName,
		event.repoName,
		event.prNumber,
		f,
	)
}

func (e *EventHandler) pullRequestOpened(
	ctx context.Context,
	event pullRequestEvent,
) (client.State, string, error) {
	cdkPath, _, target, err := e.setup(ctx, event.ownerName, event.repoName, event.prNumber, event.cloneURL)
	if err != nil {
		return client.StateError, err.Error(), err
	}
	if target == nil {
		return client.StateSuccess, "No targets are matched", nil
	}
	diff, hasDiff := e.cdk.Diff(cdkPath, "", target.Contexts)
	message := ""
	if !hasDiff {
		message = "No stacks are updated"
	}
	if err := e.cli.CreateComment(
		ctx,
		event.ownerName,
		event.repoName,
		event.prNumber,
		fmt.Sprintf("### cdk diff\n```\n%s\n```\n%s", diff, message),
	); err != nil {
		return client.StateError, err.Error(), err
	}
	if hasDiff {
		if err := e.cli.RemoveLabel(
			ctx,
			event.ownerName,
			event.repoName,
			event.prNumber,
			client.LabelNoDiff,
		); err != nil {
			return client.StateError, err.Error(), err
		}
		return client.StateFailure, "There are differences", nil
	}
	if err := e.cli.AddLabels(
		ctx,
		event.ownerName,
		event.repoName,
		event.prNumber,
		[]client.Label{client.LabelNoDiff},
	); err != nil {
		return client.StateError, err.Error(), err
	}
	return client.StateSuccess, "No diffs. Let's merge!", nil
}
