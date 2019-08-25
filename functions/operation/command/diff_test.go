package command

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/functions/operation/constant"
	"github.com/sambaiz/cdkbot/functions/operation/platform"
	"testing"

	"github.com/golang/mock/gomock"
	cdkMock "github.com/sambaiz/cdkbot/functions/operation/cdk/mock"
	"github.com/sambaiz/cdkbot/functions/operation/config"
	configMock "github.com/sambaiz/cdkbot/functions/operation/config/mock"
	gitMock "github.com/sambaiz/cdkbot/functions/operation/git/mock"
	platformMock "github.com/sambaiz/cdkbot/functions/operation/platform/mock"
	"github.com/stretchr/testify/assert"
)

func TestRunner_Diff(t *testing.T) {
	tests := []struct {
		title                  string
		cfg                    config.Config
		baseBranch             string
		resultHasDiff          bool
		resultState            constant.State
		resultStateDescription string
		isError                bool
	}{
		{
			title: "no targets are matched",
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"master": {},
				},
			},
			baseBranch:             "develop",
			resultHasDiff:          false,
			resultState:            constant.StateMergeReady,
			resultStateDescription: "No targets are matched",
		},
		{
			title: "has diffs",
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
			baseBranch:             "develop",
			resultHasDiff:          true,
			resultState:            constant.StateNeedDeploy,
			resultStateDescription: "Run /deploy after reviewed",
		},
		{
			title: "has no diffs",
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
			baseBranch:             "develop",
			resultHasDiff:          false,
			resultState:            constant.StateMergeReady,
			resultStateDescription: "No diffs. Let's merge!",
		},
	}

	constructRunnerWithMock := func(
		ctx context.Context,
		ctrl *gomock.Controller,
		cfg config.Config,
		baseBranch string,
		resultHasDiff bool,
		resultState constant.State,
		resultStateDescription string,
	) *Runner {
		platformClient := platformMock.NewMockClienter(ctrl)
		gitClient := gitMock.NewMockClienter(ctrl)
		configClient := configMock.NewMockReaderer(ctrl)
		cdkClient := cdkMock.NewMockClienter(ctrl)

		// updateStatus()
		platformClient.EXPECT().SetStatus(ctx, constant.StateRunning, "").Return(nil)
		platformClient.EXPECT().AddLabel(ctx, constant.LabelRunning).Return(nil)
		platformClient.EXPECT().SetStatus(ctx, resultState, resultStateDescription).Return(nil)
		platformClient.EXPECT().RemoveLabel(ctx, constant.LabelRunning).Return(nil)

		constructSetupMock(
			ctx,
			platformClient,
			gitClient,
			configClient,
			cdkClient,
			cfg,
			baseBranch,
		)

		target, ok := cfg.Targets[baseBranch]
		if !ok {
			return &Runner{
				platform: platformClient,
				git:      gitClient,
				config:   configClient,
				cdk:      cdkClient,
			}
		}

		platformClient.EXPECT().ListComments(ctx).Return([]platform.Comment{}, nil)
		cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
		result := "result"
		cdkClient.EXPECT().Diff(cdkPath, "", target.Contexts).Return(result, resultHasDiff)
		platformClient.EXPECT().CreateComment(ctx, fmt.Sprintf("### cdk diff\n```\n%s\n```", result)).Return(nil)
		platformClient.EXPECT().RemoveLabel(ctx, constant.LabelOutdatedDiff).Return(nil)

		return &Runner{
			platform: platformClient,
			git:      gitClient,
			config:   configClient,
			cdk:      cdkClient,
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
				test.resultState,
				test.resultStateDescription)
			assert.Equal(t, test.isError, runner.Diff(ctx) != nil)
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
			title: "Delete diff comments up to previous deploy",
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
			title: "Don't delete anything other than diff comments",
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
			title: "Don't delete diff comments before previous deploy",
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
