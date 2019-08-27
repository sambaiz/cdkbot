package command

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/functions/operation/constant"
	"strings"
)

// Rollback runs cdk deploy at base branch
func (r *Runner) Rollback(
	ctx context.Context,
	userName string,
	stacks []string,
) error {
	return r.updateStatus(ctx, func() (*resultState, error) {
		cdkPath, cfg, target, pr, err := r.setup(ctx, false)
		if err != nil {
			return nil, err
		}
		if target == nil {
			return newResultState(constant.StateMergeReady, "No targets are matched"), nil
		}
		if !cfg.IsUserAllowedDeploy(userName) {
			return newResultState(constant.StateError, fmt.Sprintf("user %s is not allowed to deploy", userName)), nil
		}
		if _, ok := pr.Labels[constant.LabelDeployed.Name]; !ok {
			return newResultState(constant.StateError, "PR is not deployed"), nil
		}
		if len(stacks) == 0 {
			stacks, err = r.cdk.List(cdkPath, target.Contexts)
			if err != nil {
				return nil, err
			}
		}
		result, err := r.cdk.Deploy(cdkPath, strings.Join(stacks, " "), target.Contexts)
		if err != nil {
			return nil, err
		}
		_, hasDiff := r.cdk.Diff(cdkPath, "", target.Contexts)
		message := "Rollback is completed."
		if hasDiff {
			message = "To be continued."
		}
		if err := r.platform.CreateComment(
			ctx,
			fmt.Sprintf("### cdk deploy (rollback)\n```\n%s\n```\n%s", result, message),
		); err != nil {
			return nil, err
		}
		if !hasDiff {
			if err := r.platform.RemoveLabel(ctx, constant.LabelDeployed); err != nil {
				return nil, err
			}
		}
		return newResultState(constant.StateNeedDeploy, "Run /deploy after reviewed"), nil
	})
}
