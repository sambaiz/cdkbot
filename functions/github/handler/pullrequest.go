package handler

import (
	"context"
	"fmt"

	"github.com/google/go-github/v26/github"
	"github.com/sambaiz/cdkbot/functions/github/client"
	"github.com/sambaiz/cdkbot/lib/cdk"
	"github.com/sambaiz/cdkbot/lib/git"
)

// PullRequestEvent handles github.PullRequestEvent
func PullRequestEvent(
	ctx context.Context,
	hook *github.PullRequestEvent,
	cli client.Clienter,
) error {
	if hook.GetAction() == "opened" {
		return pullRequestOpened(ctx, hook, cli)
	}
	return nil
}

func pullRequestOpened(
	ctx context.Context,
	hook *github.PullRequestEvent,
	cli client.Clienter,
) error {
	if err := cli.CreateStatusOfLatestCommit(
		ctx,
		hook.GetRepo().GetOwner().GetLogin(),
		hook.GetRepo().GetName(),
		hook.GetPullRequest().GetNumber(),
		client.StatePending,
	); err != nil {
		return err
	}

	hash, err := cli.GetPullRequestLatestCommitHash(
		ctx,
		hook.GetRepo().GetOwner().GetLogin(),
		hook.GetRepo().GetName(),
		hook.GetPullRequest().GetNumber(),
	)
	if err != nil {
		return err
	}
	if err := git.Clone(hook.GetRepo().GetCloneURL(), clonePath, &hash); err != nil {
		return err
	}

	if err := cdk.Setup(clonePath); err != nil {
		return err
	}

	diff, hasDiff := cdk.Diff(clonePath)
	message := ""
	if !hasDiff {
		message = "\nNo stacks are updated"
	}
	if err := cli.CreateComment(
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
	if err := cli.CreateStatusOfLatestCommit(
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
