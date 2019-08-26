package command

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/functions/operation/constant"
	"github.com/sambaiz/cdkbot/functions/operation/platform"
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
	return r.updateStatus(ctx, func() (*resultState, error) {
		cdkPath, cfg, target, pr, err := r.setup(ctx, true)
		if err != nil {
			return newResultState(constant.StateError, err.Error()), err
		}
		if target == nil {
			return newResultState(constant.StateMergeReady, "No targets are matched"), nil
		}
		if !cfg.IsUserAllowedDeploy(userName) {
			return newResultState(constant.StateError, fmt.Sprintf("user %s is not allowed to deploy", userName)), nil
		}
		openPRs, err := r.platform.GetOpenPullRequests(ctx)
		if err != nil {
			return newResultState(constant.StateError, err.Error()), err
		}
		if number, exists := existsOtherDeployedSameBasePRs(openPRs, pr); exists {
			return newResultState(
				constant.StateNeedDeploy,
				fmt.Sprintf("deplyoed PR #%d is still opened. First /deploy and merge it, or /rollback.", number),
			), nil
		}
		if len(stacks) == 0 {
			stacks, err = r.cdk.List(cdkPath, target.Contexts)
			if err != nil {
				return newResultState(constant.StateError, err.Error()), err
			}
		}
		result, err := r.cdk.Deploy(cdkPath, strings.Join(stacks, " "), target.Contexts)
		if err != nil {
			return newResultState(constant.StateError, err.Error()), err
		}
		if err := r.platform.AddLabel(ctx, constant.LabelDeployed); err != nil {
			return newResultState(constant.StateError, err.Error()), err
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
			return newResultState(constant.StateError, err.Error()), err
		}
		if !hasDiff {
			if err := r.platform.MergePullRequest(ctx, "automatically merged by cdkbot"); err != nil {
				if err := r.platform.CreateComment(
					ctx,
					fmt.Sprintf("cdkbot tried to merge but failed: %s", err.Error()),
				); err != nil {
					return newResultState(constant.StateError, err.Error()), err
				}
			} else {
				for _, openPR := range openPRs {
					if openPR.Number == pr.Number || openPR.BaseBranch != pr.BaseBranch {
						continue
					}
					if err := r.platform.AddLabelToOtherPR(ctx, constant.LabelOutdatedDiff, openPR.Number); err != nil {
						return newResultState(constant.StateError, err.Error()), err
					}
				}
			}
			return newResultState(constant.StateMergeReady, "No diffs. Let's merge!"), nil
		}
		return newResultState(constant.StateNeedDeploy, "Fix if needed and complete deploy."), nil
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

func existsOtherDeployedSameBasePRs(openPRs []platform.PullRequest, pr *platform.PullRequest) (int, bool) {
	for _, openPR := range openPRs {
		if openPR.Number == pr.Number || openPR.BaseBranch != pr.BaseBranch {
			continue
		}
		if _, ok := openPR.Labels[constant.LabelDeployed.Name]; ok {
			return openPR.Number, true
		}
	}
	return 0, false
}
