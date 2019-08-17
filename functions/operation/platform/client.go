package platform

import (
	"context"
	"github.com/sambaiz/cdkbot/functions/operation/constant"
)

// Clienter is interface of platform client
type Clienter interface {
	CreateComment(
		ctx context.Context,
		body string,
	) error
	AddLabel(
		ctx context.Context,
		label constant.Label,
	) error
	AddLabelToOtherPRs(
		ctx context.Context,
		label constant.Label,
	) error
	RemoveLabel(
		ctx context.Context,
		label constant.Label,
	) error
	GetPullRequestBaseBranch(ctx context.Context) (string, error)
	GetPullRequestLatestCommitHash(ctx context.Context) (string, error)
	GetPullRequestLabels(ctx context.Context) (map[string]constant.Label, error)
	SetStatus(
		ctx context.Context,
		state constant.State,
		description string,
	) error
}
