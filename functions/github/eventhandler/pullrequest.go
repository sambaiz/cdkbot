package eventhandler

import (
	"context"
	"fmt"

	"github.com/google/go-github/v26/github"
	"github.com/sambaiz/cdkbot/functions/github/client"
	"github.com/sambaiz/cdkbot/lib/cdk"
	"github.com/sambaiz/cdkbot/lib/config"
	"github.com/sambaiz/cdkbot/lib/git"
)

// PullRequest handles github.PullRequestEvent
func PullRequest(
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
	cfg, err := config.Read(fmt.Sprintf("%s/cdkbot.yml", clonePath))
	if err != nil {
		return err
	}
	cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)

	if err := cdk.Setup(cdkPath); err != nil {
		return err
	}
	diff, hasDiff := cdk.Diff(cdkPath)
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
