package eventhandler

import (
	"context"
	"fmt"

	"github.com/google/go-github/v26/github"
	"github.com/sambaiz/cdkbot/functions/github/client"
)

// PullRequest handles github.PullRequestEvent
func (e *EventHandler) PullRequest(
	ctx context.Context,
	hook *github.PullRequestEvent,
) error {
	if hook.GetAction() == "opened" {
		return e.pullRequestOpened(ctx, hook)
	}
	return nil
}

func (e *EventHandler) pullRequestOpened(
	ctx context.Context,
	hook *github.PullRequestEvent,
) error {
	if err := e.cli.CreateStatusOfLatestCommit(
		ctx,
		hook.GetRepo().GetOwner().GetLogin(),
		hook.GetRepo().GetName(),
		hook.GetPullRequest().GetNumber(),
		client.StatePending,
	); err != nil {
		return err
	}
	cdkPath, _, target, err := e.setup(
		ctx,
		hook.GetRepo().GetOwner().GetLogin(),
		hook.GetRepo().GetName(),
		hook.GetPullRequest().GetNumber(),
		hook.GetRepo().GetCloneURL(),
	)
	if err != nil {
		return nil
	}
	diff, hasDiff := e.cdk.Diff(cdkPath, "", target.Contexts)
	message := ""
	if !hasDiff {
		message = "\nNo stacks are updated"
	}
	if err := e.cli.CreateComment(
		ctx,
		hook.GetRepo().GetOwner().GetLogin(),
		hook.GetRepo().GetName(),
		hook.GetPullRequest().GetNumber(),
		fmt.Sprintf("### cdk diff\n```%s```\n%s", diff, message),
	); err != nil {
		return err
	}
	status := client.StateSuccess
	if hasDiff {
		status = client.StateFailure
	}
	if err := e.cli.CreateStatusOfLatestCommit(
		ctx,
		hook.GetRepo().GetOwner().GetLogin(),
		hook.GetRepo().GetName(),
		hook.GetPullRequest().GetNumber(),
		status,
	); err != nil {
		return err
	}
	return nil
}
