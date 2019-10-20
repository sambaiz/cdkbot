package command

import (
	"context"
	"fmt"
	cdkMock "github.com/sambaiz/cdkbot/tasks/operation/cdk/mock"
	"github.com/sambaiz/cdkbot/tasks/operation/config"
	configMock "github.com/sambaiz/cdkbot/tasks/operation/config/mock"
	"github.com/sambaiz/cdkbot/tasks/operation/constant"
	gitMock "github.com/sambaiz/cdkbot/tasks/operation/git/mock"
	"github.com/sambaiz/cdkbot/tasks/operation/logger"
	"github.com/sambaiz/cdkbot/tasks/operation/platform"
	platformMock "github.com/sambaiz/cdkbot/tasks/operation/platform/mock"
	"testing"

	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRunner_Diff(t *testing.T) {
	type expected struct {
		outState *resultState
		isError  bool
	}
	tests := []struct {
		title         string
		cfg           config.Config
		baseBranch    string
		resultHasDiff bool
		diffError     error
		expected      expected
	}{
		{
			title: "no_targets_are_matched",
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"master": {},
				},
			},
			baseBranch:    "develop",
			resultHasDiff: false,
			expected: expected{
				outState: newResultState(constant.StateMergeReady, "No targets are matched"),
				isError:  false,
			},
		},
		{
			title: "has_diffs",
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"develop": {
						Contexts: map[string]string{
							"env": "stg",
						},
					},
				},
			},
			baseBranch:    "develop",
			resultHasDiff: true,
			expected: expected{
				outState: newResultState(constant.StateNotMergeReady, "Run /deploy after reviewed"),
				isError:  false,
			},
		},
		{
			title: "cdk_diff_error",
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"develop": {
						Contexts: map[string]string{
							"env": "stg",
						},
					},
				},
			},
			baseBranch: "develop",
			diffError:  errors.New("cdk diff error"),
			expected: expected{
				outState: newResultState(constant.StateNotMergeReady, "Fix codes"),
				isError:  false,
			},
		},
		{
			title: "has_no_diffs",
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"develop": {
						Contexts: map[string]string{
							"env": "stg",
						},
					},
				},
			},
			baseBranch:    "develop",
			resultHasDiff: false,
			expected: expected{
				outState: newResultState(constant.StateMergeReady, "No diffs. Let's merge!"),
				isError:  false,
			},
		},
	}

	constructRunnerWithMock := func(
		ctx context.Context,
		ctrl *gomock.Controller,
		cfg config.Config,
		baseBranch string,
		resultHasDiff bool,
		diffError error,
		expected expected,
	) *Runner {
		platformClient := platformMock.NewMockClienter(ctrl)
		gitClient := gitMock.NewMockClienter(ctrl)
		configClient := configMock.NewMockReaderer(ctrl)
		cdkClient := cdkMock.NewMockClienter(ctrl)

		// updateStatus()
		platformClient.EXPECT().SetStatus(ctx, constant.StateRunning, "").Return(nil)
		platformClient.EXPECT().AddLabel(ctx, constant.LabelRunning).Return(nil)
		platformClient.EXPECT().SetStatus(ctx, expected.outState.state, expected.outState.description).Return(nil)
		platformClient.EXPECT().RemoveLabel(ctx, constant.LabelRunning).Return(nil)

		constructSetupMock(
			ctx,
			platformClient,
			gitClient,
			configClient,
			cdkClient,
			true,
			cfg,
			&platform.PullRequest{
				BaseBranch:     baseBranch,
				BaseCommitHash: "basehash",
				HeadCommitHash: "headhash",
				Labels:         nil,
			},
		)

		target, ok := cfg.Targets[baseBranch]
		if !ok {
			return &Runner{
				platform: platformClient,
				git:      gitClient,
				config:   configClient,
				cdk:      cdkClient,
				logger:   logger.MockLogger{},
			}
		}

		platformClient.EXPECT().ListComments(ctx).Return([]platform.Comment{}, nil)
		cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
		result := "result"
		cdkClient.EXPECT().Diff(cdkPath, "", target.Contexts).Return(result, resultHasDiff, diffError)
		platformClient.EXPECT().CreateComment(ctx, fmt.Sprintf("### cdk diff\n```\n%s\n```", result)).Return(nil)
		if diffError != nil {
			return &Runner{
				platform: platformClient,
				git:      gitClient,
				config:   configClient,
				cdk:      cdkClient,
				logger:   logger.MockLogger{},
			}
		}
		platformClient.EXPECT().RemoveLabel(ctx, constant.LabelOutdatedDiff).Return(nil)

		return &Runner{
			platform: platformClient,
			git:      gitClient,
			config:   configClient,
			cdk:      cdkClient,
			logger:   logger.MockLogger{},
		}
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			runner := constructRunnerWithMock(
				ctx,
				ctrl,
				test.cfg,
				test.baseBranch,
				test.resultHasDiff,
				test.diffError,
				test.expected)
			assert.Equal(t, test.expected.isError, runner.Diff(ctx) != nil)
		})
	}
}

func TestRunner_deleteDiffCommentsUpToPreviousDeploy(t *testing.T) {
	diffComment := "### cdk diff\n```\nresult\n```"
	deployComment := "### cdk deploy\n```\nresult\n```"
	tests := []struct {
		title              string
		in                 []platform.Comment
		expectedDeletedIDs []int64
		isError            bool
	}{
		{
			title: "Delete_diff_comments_up_to_previous_deploy",
			in: []platform.Comment{
				{
					ID:   1,
					Body: diffComment,
				},
				{
					ID:   2,
					Body: diffComment,
				},
			},
			expectedDeletedIDs: []int64{1, 2},
		},
		{
			title: "Don't_delete_anything_other_than_diff_comments",
			in: []platform.Comment{
				{
					ID:   1,
					Body: "a",
				},
				{
					ID:   2,
					Body: diffComment,
				},
			},
			expectedDeletedIDs: []int64{2},
		},
		{
			title: "Don't_delete_diff_comments_before_previous_deploy",
			in: []platform.Comment{
				{
					ID:   1,
					Body: diffComment,
				},
				{
					ID:   2,
					Body: deployComment,
				},
			},
			expectedDeletedIDs: []int64{},
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			platformClient := platformMock.NewMockClienter(ctrl)
			for _, id := range test.expectedDeletedIDs {
				platformClient.EXPECT().DeleteComment(ctx, id).Return(nil)
			}
			client := &Runner{
				platform: platformClient,
			}
			err := client.deleteDiffCommentsUpToPreviousDeploy(ctx, test.in)
			assert.Equal(t, test.isError, err != nil)
		})
	}
}
