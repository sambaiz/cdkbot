package eventhandler

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v26/github"
	"github.com/sambaiz/cdkbot/functions/github/client"
)

const clonePath = "/tmp/repo"

type issueCommentEvent struct {
	ownerName   string
	repoName    string
	issueNumber int
	commentBody string
	cloneURL    string
}

// IssueComment handles
func (e *EventHandler) IssueComment(
	ctx context.Context,
	hook *github.IssueCommentEvent,
) error {
	event := issueCommentEvent{
		ownerName:   hook.GetRepo().GetOwner().GetLogin(),
		repoName:    hook.GetRepo().GetName(),
		issueNumber: hook.GetIssue().GetNumber(),
		commentBody: hook.GetComment().GetBody(),
		cloneURL:    hook.GetRepo().GetCloneURL(),
	}
	if hook.GetAction() == "created" {
		return e.issueCommentCreated(ctx, event)
	}
	return nil
}

func (e *EventHandler) issueCommentCreated(
	ctx context.Context,
	event issueCommentEvent,
) error {
	cmd := parseCommand(event.commentBody)
	if cmd == nil {
		return nil
	}
	if err := e.cli.CreateStatusOfLatestCommit(
		ctx,
		event.ownerName,
		event.repoName,
		event.issueNumber,
		client.StatePending,
	); err != nil {
		return err
	}
	hash, err := e.cli.GetPullRequestLatestCommitHash(
		ctx,
		event.ownerName,
		event.repoName,
		event.issueNumber,
	)
	if err != nil {
		return err
	}
	if err := e.git.Clone(event.cloneURL, clonePath, &hash); err != nil {
		return err
	}
	cfg, err := e.config.Read(fmt.Sprintf("%s/cdkbot.yml", clonePath))
	if err != nil {
		return err
	}
	cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
	baseBranch, err := e.cli.GetPullRequestBaseBranch(
		ctx,
		event.ownerName,
		event.repoName,
		event.issueNumber,
	)
	if err != nil {
		return err
	}
	if _, ok := cfg.Targets[baseBranch]; !ok {
		// noop
		return nil
	}

	if err := e.cdk.Setup(cdkPath); err != nil {
		return err
	}

	switch cmd.action {
	case actionDiff:
		err = e.doActionDiff(ctx, event, cdkPath, cmd.args)
	case actionDeploy:
		err = e.doActionDeploy(ctx, event, cdkPath, cmd.args)
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

func parseCommand(cmd string) *command {
	if !strings.HasPrefix(cmd, "/") {
		return nil
	}
	parts := strings.Split(cmd, " ")
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
	event issueCommentEvent,
	cdkPath string,
	cmdArgs string,
) error {
	diff, hasDiff := e.cdk.Diff(cdkPath)
	if err := e.cli.CreateComment(
		ctx,
		event.ownerName,
		event.repoName,
		event.issueNumber,
		fmt.Sprintf("### cdk diff %s\n```%s```", strings.TrimSpace(cmdArgs), diff),
	); err != nil {
		return err
	}
	status := client.StateSuccess
	if hasDiff {
		status = client.StateFailure
	}
	if err := e.cli.CreateStatusOfLatestCommit(
		ctx,
		event.ownerName,
		event.repoName,
		event.issueNumber,
		status,
	); err != nil {
		return err
	}
	return nil
}

func (e *EventHandler) doActionDeploy(
	ctx context.Context,
	event issueCommentEvent,
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
		event.ownerName,
		event.repoName,
		event.issueNumber,
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
		event.ownerName,
		event.repoName,
		event.issueNumber,
		status,
	); err != nil {
		return err
	}
	return nil
}
