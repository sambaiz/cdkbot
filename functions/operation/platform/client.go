package platform

import (
	"context"

	"github.com/sambaiz/cdkbot/functions/operation/constant"
)

// Comment is a comment of PR
type Comment struct {
	ID   int64
	Body string
}

// Clienter is interface of platform client
type Clienter interface {
	CreateComment(
		ctx context.Context,
		body string,
	) error
	ListComments(
		ctx context.Context,
	) ([]Comment, error)
	DeleteComment(
		ctx context.Context,
		commentID int64,
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
	GetPullRequestCommitHash(ctx context.Context) (string, string, error)
	GetPullRequestLabels(ctx context.Context) (map[string]constant.Label, error)
	GetOpenPullRequestNumbersByLabel(
		ctx context.Context,
		label constant.Label,
		excludeMySelf bool,
	) ([]int, error)
	MergePullRequest(ctx context.Context, message string) error
	SetStatus(
		ctx context.Context,
		state constant.State,
		description string,
	) error
}
