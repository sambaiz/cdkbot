package command

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/functions/operation/constant"
	"github.com/sambaiz/cdkbot/functions/operation/platform"
	"strings"
)

// Diff runs cdk diff
func (r *Runner) Diff(
	ctx context.Context,
) error {
	return r.updateStatus(ctx, func() (*resultState, error) {
		cdkPath, _, target, _, err := r.setup(ctx, true)
		if err != nil {
			return nil, err
		}
		if target == nil {
			return newResultState(constant.StateMergeReady, "No targets are matched"), nil
		}

		comments, err := r.platform.ListComments(ctx)
		if err != nil {
			return nil, err
		}
		diff, hasDiff, diffErr := r.cdk.Diff(cdkPath, "", target.Contexts)
		if err := r.platform.CreateComment(
			ctx,
			fmt.Sprintf("### cdk diff\n```\n%s\n```", diff),
		); err != nil {
			return nil, err
		}
		// Leave only one diff comment after previous deploy to clean PR
		if err := r.deleteDiffCommentsUpToPreviousDeploy(ctx, comments); err != nil {
			return nil, err
		}
		if diffErr != nil {
			return newResultState(constant.StateNeedDeploy, "Fix codes"), nil
		}
		if err := r.platform.RemoveLabel(ctx, constant.LabelOutdatedDiff); err != nil {
			return nil, err
		}
		if hasDiff {
			return newResultState(constant.StateNeedDeploy, "Run /deploy after reviewed"), nil
		}
		return newResultState(constant.StateMergeReady, "No diffs. Let's merge!"), nil
	})
}

func (r *Runner) deleteDiffCommentsUpToPreviousDeploy(ctx context.Context, comments []platform.Comment) error {
	for i := len(comments) - 1; i >= 0; i-- {
		if strings.HasPrefix(comments[i].Body, "### cdk deploy\n") {
			return nil
		}
		if strings.HasPrefix(comments[i].Body, "### cdk diff\n") {
			if err := r.platform.DeleteComment(ctx, comments[i].ID); err != nil {
				return err
			}
		}
	}
	return nil
}
