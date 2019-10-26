package platform

import (
	"context"
	"github.com/sambaiz/cdkbot/tasks/operation/constant"
)

// Comment is a comment of PR
type Comment struct {
	ID   int64
	Body string
}

// PullRequest is a PR
type PullRequest struct {
	Number         int
	BaseBranch     string
	BaseCommitHash string
	HeadCommitHash string
	Labels         map[string]constant.Label
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
	AddLabelToOtherPR(
		ctx context.Context,
		label constant.Label,
		number int,
	) error
	RemoveLabel(
		ctx context.Context,
		label constant.Label,
	) error
	GetPullRequest(ctx context.Context) (*PullRequest, error)
	GetOpenPullRequests(ctx context.Context) ([]PullRequest, error)
	MergePullRequest(ctx context.Context, message string) error
	SetStatus(
		ctx context.Context,
		state constant.State,
		description string,
	) error
}
