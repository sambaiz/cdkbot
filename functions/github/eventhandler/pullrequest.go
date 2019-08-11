package eventhandler

import (
	"context"
	"fmt"

	"github.com/google/go-github/v26/github"
	"github.com/sambaiz/cdkbot/functions/github/client"
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
	var f func() (client.State, error)
	switch ev.GetAction() {
	case "opened":
		f = func() (client.State, error) {
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
) (client.State, error) {
	cdkPath, _, target, err := e.setup(ctx, event.ownerName, event.repoName, event.prNumber, event.cloneURL)
	if err != nil {
		return client.StateError, err
	}
	if target == nil {
		return client.StateSuccess, err
	}
	diff, hasDiff := e.cdk.Diff(cdkPath, "", target.Contexts)
	message := ""
	if !hasDiff {
		message = "\nNo stacks are updated"
	}
	if err := e.cli.CreateComment(
		ctx,
		event.ownerName,
		event.repoName,
		event.prNumber,
		fmt.Sprintf("### cdk diff\n```%s```\n%s", diff, message),
	); err != nil {
		return client.StateError, err
	}
	if hasDiff {
		return client.StateFailure, nil
	}
	return client.StateSuccess, nil
}
