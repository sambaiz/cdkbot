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
	var f func() (client.State, error)
	switch hook.GetAction() {
	case "created":
		f = func() (client.State, error) {
			return e.issueCommentCreated(ctx, event)
		}
	default:
		return nil
	}
	return e.updateStatus(
		ctx,
		event.ownerName,
		event.repoName,
		event.issueNumber,
		f,
	)
}

func (e *EventHandler) issueCommentCreated(
	ctx context.Context,
	event issueCommentEvent,
) (client.State, error) {
	cmd := parseCommand(event.commentBody)
	if cmd == nil {
		return client.StateSuccess, nil
	}
	cdkPath, _, target, err := e.setup(
		ctx,
		event.ownerName,
		event.repoName,
		event.issueNumber,
		event.cloneURL,
	)
	if err != nil {
		return client.StateError, err
	}
	if target == nil {
		return client.StateSuccess, err
	}
	var hasDiff bool
	switch cmd.action {
	case actionDiff:
		hasDiff, err = e.doActionDiff(ctx, event, cdkPath, cmd.args, target.Contexts)
	case actionDeploy:
		hasDiff, err = e.doActionDeploy(ctx, event, cdkPath, cmd.args, target.Contexts)
	default:
		return client.StateSuccess, nil
	}
	if err != nil {
		return client.StateError, err
	}
	if hasDiff {
		return client.StateFailure, nil
	} else {
		return client.StateSuccess, nil
	}
	return client.StateSuccess, nil
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
	contexts map[string]string,
) (bool, error) {
	args := strings.TrimSpace(strings.Replace(cmdArgs, "\n", " ", -1))
	diff, hasDiff := e.cdk.Diff(cdkPath, args, contexts)
	if err := e.cli.CreateComment(
		ctx,
		event.ownerName,
		event.repoName,
		event.issueNumber,
		fmt.Sprintf("### cdk diff %s\n```%s```", args, diff),
	); err != nil {
		return false, err
	}
	return hasDiff, nil
}

func (e *EventHandler) doActionDeploy(
	ctx context.Context,
	event issueCommentEvent,
	cdkPath string,
	cmdArgs string,
	contexts map[string]string,
) (bool, error) {
	args := strings.TrimSpace(strings.Replace(cmdArgs, "\n", " ", -1))
	if len(args) == 0 {
		stacks, err := e.cdk.List(cdkPath, contexts)
		if err != nil {
			return false, err
		}
		args = strings.Join(stacks, " ")
	}
	result, err := e.cdk.Deploy(cdkPath, args, contexts)
	if err != nil {
		return false, err
	}
	_, hasDiff := e.cdk.Diff(cdkPath, "", contexts)
	message := "All stacks have been deployed :tada:"
	if hasDiff {
		message = "To be continued"
	}
	if err := e.cli.CreateComment(
		ctx,
		event.ownerName,
		event.repoName,
		event.issueNumber,
		fmt.Sprintf("### cdk deploy %s\n```%s```\n%s", args, result, message),
	); err != nil {
		return false, err
	}
	return hasDiff, nil
}
