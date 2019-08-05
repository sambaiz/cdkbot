package eventhandler

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v26/github"
	"github.com/sambaiz/cdkbot/functions/github/client"
)

const clonePath = "/tmp/repo"

// IssueComment handles github.IssueCommentEvent
func (e *EventHandler) IssueComment(
	ctx context.Context,
	hook *github.IssueCommentEvent,
) error {
	if hook.GetAction() == "created" {
		return e.issueCreated(ctx, hook)
	}
	return nil
}

func (e *EventHandler) issueCreated(
	ctx context.Context,
	hook *github.IssueCommentEvent,
) error {
	cmd := parseCommand(hook.GetComment().GetBody())
	if cmd == nil {
		return nil
	}
	if err := e.cli.CreateStatusOfLatestCommit(
		ctx,
		hook.GetRepo().GetOwner().GetLogin(),
		hook.GetRepo().GetName(),
		hook.GetIssue().GetNumber(),
		client.StatePending,
	); err != nil {
		return err
	}
	hash, err := e.cli.GetPullRequestLatestCommitHash(
		ctx,
		hook.GetRepo().GetOwner().GetLogin(),
		hook.GetRepo().GetName(),
		hook.GetIssue().GetNumber(),
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
	cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)

	if err := e.cdk.Setup(cdkPath); err != nil {
		return err
	}

	switch cmd.action {
	case actionDiff:
		err = e.doActionDiff(ctx, hook, cdkPath, cmd.args)
	case actionDeploy:
		err = e.doActionDeploy(ctx, hook, cdkPath, cmd.args)
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

func (e *EventHandler) doActionDiff(
	ctx context.Context,
	hook *github.IssueCommentEvent,
	cdkPath string,
	cmdArgs string,
) error {
	diff, _ := e.cdk.Diff(cdkPath)
	if err := e.cli.CreateComment(
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

func (e *EventHandler) doActionDeploy(
	ctx context.Context,
	hook *github.IssueCommentEvent,
	cdkPath string,
	cmdArgs string,
) error {
	args := strings.TrimSpace(strings.Replace(cmdArgs, "\n", " ", -1))
	if len(args) == 0 {
		stacks, err := e.cdk.List(cdkPath)
		if err != nil {
			return err
		}
		args = strings.Join(stacks, " ")
	}
	result, err := e.cdk.Deploy(cdkPath, args)
	if err != nil {
		return err
	}
	_, hasDiff := e.cdk.Diff(cdkPath)
	message := "All stacks have been deployed :tada:"
	if hasDiff {
		message = "To be continued"
	}
	if err := e.cli.CreateComment(
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
	if err := e.cli.CreateStatusOfLatestCommit(
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
