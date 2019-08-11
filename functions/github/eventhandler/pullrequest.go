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

	hash, err := e.cli.GetPullRequestLatestCommitHash(
		ctx,
		hook.GetRepo().GetOwner().GetLogin(),
		hook.GetRepo().GetName(),
		hook.GetPullRequest().GetNumber(),
	)
	if err != nil {
		return err
	}
	if err := e.git.Clone(hook.GetRepo().GetCloneURL(), clonePath, &hash); err != nil {
		return err
	}
	cfg, err := e.config.Read(fmt.Sprintf("%s/cdkbot.yml", clonePath))
	if err != nil {
		return err
	}
	target, ok := cfg.Targets[hook.GetPullRequest().GetBase().GetLabel()]
	if !ok {
		// noop
		return nil
	}

	cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)

	if err := e.cdk.Setup(cdkPath); err != nil {
		return err
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
