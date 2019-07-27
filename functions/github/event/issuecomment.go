package event

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v26/github"
	"github.com/sambaiz/cdkbot/functions/github/client"
	"github.com/sambaiz/cdkbot/lib/cdk"
	"github.com/sambaiz/cdkbot/lib/git"
)

const clonePath = "/tmp/repo"

// HandleIssueCommentEvent handles IssueCommentEvent
func HandleIssueCommentEvent(
	ctx context.Context,
	hook *github.IssueCommentEvent,
	cli client.Clienter,
) error {
	if hook.GetAction() == "created" {
		return handleIssueCreated(ctx, hook, cli)
	}
	return nil
}

func handleIssueCreated(
	ctx context.Context,
	hook *github.IssueCommentEvent,
	cli client.Clienter,
) error {
	cmd := parseCommand(hook.GetComment().GetBody())
	if cmd == nil {
		return nil
	}
	if err := cli.CreateStatusOfLatestCommit(
		ctx,
		hook.GetRepo().GetOwner().GetLogin(),
		hook.GetRepo().GetName(),
		hook.GetIssue().GetNumber(),
		client.StatePending,
	); err != nil {
		return err
	}
	hash, err := cli.GetPullRequestLatestCommitHash(
		ctx,
		hook.GetRepo().GetOwner().GetLogin(),
		hook.GetRepo().GetName(),
		hook.GetIssue().GetNumber(),
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

	switch cmd.action {
	case actionDiff:
		err = doActionDiff(ctx, hook, cli, cmd.args)
	case actionDeploy:
		err = doActionDeploy(ctx, hook, cli, cmd.args)
	}
	return err
}

type action string

var (
	actionDiff   action = "diff"
	actionDeploy action = "deploy"
)

type command struct {
	action action
	args   string
}

func parseCommand(comment string) *command {
	if !strings.HasPrefix(comment, "/") {
		return nil
	}
	parts := strings.Split(comment, " ")
	switch parts[0] {
	case "/deploy":
		return &command{
			action: actionDeploy,
			args:   strings.Join(parts[1:], " "),
		}
	case "/diff":
		return &command{
			action: actionDiff,
			args:   strings.Join(parts[1:], " "),
		}
	}
	return nil
}

func doActionDiff(
	ctx context.Context,
	hook *github.IssueCommentEvent,
	cli client.Clienter,
	cmdArgs string,
) error {
	diff, _ := cdk.Diff(clonePath)
	if err := cli.CreateComment(
		ctx,
		hook.GetRepo().GetOwner().GetLogin(),
		hook.GetRepo().GetName(),
		hook.GetIssue().GetNumber(),
		fmt.Sprintf("### cdk diff %s\n```%s```", strings.TrimSpace(cmdArgs), diff),
	); err != nil {
		return err
	}
	return nil
}

func doActionDeploy(
	ctx context.Context,
	hook *github.IssueCommentEvent,
	cli client.Clienter,
	cmdArgs string,
) error {
	args := strings.TrimSpace(strings.Replace(cmdArgs, "\n", " ", -1))
	if len(args) == 0 {
		stacks, err := cdk.List(clonePath)
		if err != nil {
			return err
		}
		args = strings.Join(stacks, " ")
	}
	result, err := cdk.Deploy(clonePath, args)
	if err != nil {
		return err
	}
	_, hasDiff := cdk.Diff(clonePath)
	message := "All stacks have been deployed :tada:"
	if hasDiff {
		message = "To be continued"
	}
	if err := cli.CreateComment(
		ctx,
		hook.GetRepo().GetOwner().GetLogin(),
		hook.GetRepo().GetName(),
		hook.GetIssue().GetNumber(),
		fmt.Sprintf("### cdk deploy\n```%s```\n%s", result, message),
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
		hook.GetIssue().GetNumber(),
		status,
	); err != nil {
		return err
	}
	return nil
}
