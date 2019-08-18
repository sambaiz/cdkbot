package eventhandler

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/functions/operation/constant"
	"strings"
)

// CommentCreated handles the event of comment created
func (e *EventHandler) CommentCreated(
	ctx context.Context,
	userName string,
	comment string,
) error {
	cmd := parseCommand(comment)
	if cmd == nil {
		return nil
	}
	return e.updateStatus(ctx, func() (constant.State, string, error) {
		cdkPath, cfg, target, err := e.setup(ctx)
		if err != nil {
			return constant.StateError, err.Error(), err
		}
		if target == nil {
			return constant.StateMergeReady, "No targets are matched", nil
		}
		if cmd.action == actionDeploy {
			if has, _ := e.hasOutdatedDiffLabel(ctx); has {
				if err := e.platform.CreateComment(ctx, "Differences are outdated. Run /diff instead."); err != nil {
					return constant.StateError, err.Error(), err
				}
				cmd.action = actionDiff
				cmd.args = ""
			}
		}
		var hasDiff bool
		switch cmd.action {
		case actionDiff:
			hasDiff, err = e.doActionDiff(ctx, cdkPath, cmd.args, target.Contexts)
			if err != nil {
				return constant.StateError, err.Error(), err
			}
		case actionDeploy:
			if !cfg.IsUserAllowedDeploy(userName) {
				return constant.StateError, fmt.Sprintf("User %s is not allowed to deploy", userName), nil
			}
			hasDiff, err = e.doActionDeploy(ctx, cdkPath, cmd.args, target.Contexts)
			if err != nil {
				return constant.StateError, err.Error(), err
			}
		default:
			return constant.StateError, fmt.Sprintf("Command %s is unknown", cmd.action), nil
		}
		if hasDiff {
			if err := e.platform.RemoveLabel(ctx, constant.LabelNoDiff); err != nil {
				return constant.StateError, err.Error(), err
			}
			return constant.StateNeedDeploy, "Diffs still remain", nil
		}
		if err := e.platform.AddLabel(ctx, constant.LabelNoDiff); err != nil {
			return constant.StateError, err.Error(), err
		}
		return constant.StateMergeReady, "No diffs. Let's merge!", nil
	})
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

func (e *EventHandler) hasOutdatedDiffLabel(ctx context.Context) (bool, error) {
	// get labels from not event but API because to get latest one.
	labels, err := e.platform.GetPullRequestLabels(ctx)
	if err != nil {
		return false, err
	}
	if _, ok := labels[constant.LabelOutdatedDiff.Name]; ok {
		return true, nil
	}
	return false, nil
}

func (e *EventHandler) doActionDiff(
	ctx context.Context,
	cdkPath string,
	cmdArgs string,
	contexts map[string]string,
) (bool, error) {
	args := strings.TrimSpace(strings.Replace(cmdArgs, "\n", " ", -1))
	diff, hasDiff := e.cdk.Diff(cdkPath, args, contexts)
	if err := e.platform.CreateComment(
		ctx,
		fmt.Sprintf("### cdk diff %s\n```\n%s\n```", args, diff),
	); err != nil {
		return false, err
	}
	if err := e.platform.RemoveLabel(ctx, constant.LabelOutdatedDiff); err != nil {
		return false, err
	}
	return hasDiff, nil
}

func (e *EventHandler) doActionDeploy(
	ctx context.Context,
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
	if err := e.platform.AddLabelToOtherPRs(ctx, constant.LabelOutdatedDiff); err != nil {
		return false, err
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
	if err := e.platform.CreateComment(
		ctx,
		fmt.Sprintf("### cdk deploy %s\n```\n%s\n```\n%s", args, result, message),
	); err != nil {
		return false, err
	}
	return hasDiff, nil
}
