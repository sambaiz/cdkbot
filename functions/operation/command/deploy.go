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
	stacks []string,
) error {
	if has, _ := r.hasOutdatedDiffLabel(ctx); has {
		if err := r.platform.CreateComment(ctx, "Differences are outdated. Run /diff instead."); err != nil {
			return err
		}
		return r.Diff(ctx)
	}
	return r.updateStatus(ctx, func() (constant.State, string, error) {
		cdkPath, cfg, target, err := r.setup(ctx, true)
		if err != nil {
			return constant.StateError, err.Error(), err
		}
		if target == nil {
			return constant.StateMergeReady, "No targets are matched", nil
		}
		if !cfg.IsUserAllowedDeploy(userName) {
			return constant.StateError, fmt.Sprintf("user %s is not allowed to deploy", userName), nil
		}
		deployedPRs, err := r.platform.GetOpenPullRequestNumbersByLabel(ctx, constant.LabelDeployed, true)
		if err != nil {
			return constant.StateError, err.Error(), err
		}
		if len(deployedPRs) > 0 {
			return constant.StateNeedDeploy, fmt.Sprintf("deplyoed PR #%d is still opened. First /deploy and merge it, or /rollback.", deployedPRs[0]), nil
		}
		if len(stacks) == 0 {
			stacks, err = r.cdk.List(cdkPath, target.Contexts)
			if err != nil {
				return constant.StateError, err.Error(), err
			}
		}
		result, err := r.cdk.Deploy(cdkPath, strings.Join(stacks, " "), target.Contexts)
		if err != nil {
			return constant.StateError, err.Error(), err
		}
		if err := r.platform.AddLabel(ctx, constant.LabelDeployed); err != nil {
			return constant.StateError, err.Error(), err
		}
		_, hasDiff := r.cdk.Diff(cdkPath, "", target.Contexts)
		message := "Success :tada:"
		if hasDiff {
			message = "To be continued."
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
			} else {
				if err := r.platform.AddLabelToOtherPRs(ctx, constant.LabelOutdatedDiff); err != nil {
					return constant.StateError, err.Error(), err
				}
			}
			return constant.StateMergeReady, "No diffs. Let's merge!", nil
		}
		return constant.StateNeedDeploy, "Fix if needed and complete deploy.", nil
	})
}

func (r *Runner) hasOutdatedDiffLabel(ctx context.Context) (bool, error) {
	// get labels from not event but API because to get latest data.
	pr, err := r.platform.GetPullRequest(ctx)
	if err != nil {
		return false, err
	}
	if _, ok := pr.Labels[constant.LabelOutdatedDiff.Name]; ok {
		return true, nil
	}
	return false, nil
}
