package command

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/functions/operation/constant"
	"strings"
)

// Deploy runs cdk deploy
func (r *Runner) Deploy(
	ctx context.Context,
	userName string,
) error {
	if has, _ := r.hasOutdatedDiffLabel(ctx); has {
		if err := r.platform.CreateComment(ctx, "Differences are outdated. Run /diff instead."); err != nil {
			return err
		}
		return r.Diff(ctx)
	}
	return r.updateStatus(ctx, func() (constant.State, string, error) {
		cdkPath, cfg, target, err := r.setup(ctx)
		if err != nil {
			return constant.StateError, err.Error(), err
		}
		if target == nil {
			return constant.StateMergeReady, "No targets are matched", nil
		}
		if !cfg.IsUserAllowedDeploy(userName) {
			return constant.StateError, fmt.Sprintf("user %s is not allowed to deploy", userName), nil
		}
		if err := r.platform.AddLabelToOtherPRs(ctx, constant.LabelOutdatedDiff); err != nil {
			return constant.StateError, err.Error(), err
		}
		stacks, err := r.cdk.List(cdkPath, target.Contexts)
		if err != nil {
			return constant.StateError, err.Error(), err
		}
		result, err := r.cdk.Deploy(cdkPath, strings.Join(stacks, " "), target.Contexts)
		if err != nil {
			return constant.StateError, err.Error(), err
		}
		_, hasDiff := r.cdk.Diff(cdkPath, "", target.Contexts)
		message := "All stacks have been deployed :tada:"
		if hasDiff {
			message = "Some stacks are failed to deploy... Don't give up!"
		}
		if err := r.platform.CreateComment(
			ctx,
			fmt.Sprintf("### cdk deploy\n```\n%s\n```\n%s", result, message),
		); err != nil {
			return constant.StateError, err.Error(), err
		}
		if !hasDiff {
			if err := r.platform.MergePullRequest(ctx, "automatically merged by cdkbot"); err != nil {
				if err := r.platform.CreateComment(
					ctx,
					fmt.Sprintf("cdkbot tried to merge but failed: %s", err.Error()),
				); err != nil {
					return constant.StateError, err.Error(), err
				}
			}
		}
		return constant.StateMergeReady, "No diffs. Let's merge!", nil
	})
}


func (r *Runner) hasOutdatedDiffLabel(ctx context.Context) (bool, error) {
	// get labels from not event but API because to get latest one.
	labels, err := r.platform.GetPullRequestLabels(ctx)
	if err != nil {
		return false, err
	}
	if _, ok := labels[constant.LabelOutdatedDiff.Name]; ok {
		return true, nil
	}
	return false, nil
}