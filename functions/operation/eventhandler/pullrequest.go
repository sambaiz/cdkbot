package eventhandler

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/functions/operation/constant"
)

// PullRequestOpened handles the event of PR opened
func (e *EventHandler) PullRequestOpened(ctx context.Context) error {
	return e.updateStatus(ctx, func() (constant.State, string, error) {
		cdkPath, _, target, err := e.setup(ctx)
		if err != nil {
			return constant.StateError, err.Error(), err
		}
		if target == nil {
			return constant.StateMergeReady, "No targets are matched", nil
		}
		diff, hasDiff := e.cdk.Diff(cdkPath, "", target.Contexts)
		if err := e.platform.CreateComment(
			ctx,
			fmt.Sprintf("### cdk diff\n```\n%s\n```", diff),
		); err != nil {
			return constant.StateError, err.Error(), err
		}
		if hasDiff {
			if err := e.platform.RemoveLabel(ctx, constant.LabelNoDiff); err != nil {
				return constant.StateError, err.Error(), err
			}
			return constant.StateNeedDeploy, "There are differences", nil
		}
		if err := e.platform.AddLabel(ctx, constant.LabelNoDiff); err != nil {
			return constant.StateError, err.Error(), err
		}
		return constant.StateMergeReady, "No diffs. Let's merge!", nil
	})
}
