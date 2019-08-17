package eventhandler

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/functions/operation/constant"
	"testing"

	"github.com/golang/mock/gomock"
	cdkMock "github.com/sambaiz/cdkbot/functions/operation/cdk/mock"
	"github.com/sambaiz/cdkbot/functions/operation/config"
	configMock "github.com/sambaiz/cdkbot/functions/operation/config/mock"
	gitMock "github.com/sambaiz/cdkbot/functions/operation/git/mock"
	platformMock "github.com/sambaiz/cdkbot/functions/operation/platform/mock"
	"github.com/stretchr/testify/assert"
)

func TestEventHandlerPullRequestOpened(t *testing.T) {
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
			title: "target is matched and has diffs",
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
			resultStateDescription: "There are differences",
		},
	}

	constructEventHandlerWithMock := func(
		ctx context.Context,
		ctrl *gomock.Controller,
		cfg config.Config,
		baseBranch string,
		resultHasDiff bool,
		resultState constant.State,
		resultStateDescription string,
	) *EventHandler {
		platformClient := platformMock.NewMockClienter(ctrl)
		gitClient := gitMock.NewMockClienter(ctrl)
		configClient := configMock.NewMockReaderer(ctrl)
		cdkClient := cdkMock.NewMockClienter(ctrl)

		// updateStatus()
		platformClient.EXPECT().SetStatus(ctx, constant.StateRunning, "").Return(nil)
		platformClient.EXPECT().SetStatus(ctx, resultState, resultStateDescription).Return(nil)

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
			return &EventHandler{
				platform: platformClient,
				git:      gitClient,
				config:   configClient,
				cdk:      cdkClient,
			}
		}
		cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
		result := "result"
		cdkClient.EXPECT().Diff(cdkPath, "", target.Contexts).Return(result, resultHasDiff)
		platformClient.EXPECT().CreateComment(
			ctx,
			fmt.Sprintf("### cdk diff\n```\n%s\n```", result),
		).Return(nil)

		if resultHasDiff {
			platformClient.EXPECT().RemoveLabel(ctx, constant.LabelNoDiff).Return(nil)
		} else {
			platformClient.EXPECT().AddLabel(ctx, constant.LabelNoDiff).Return(nil)
		}

		return &EventHandler{
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
			eventHandler := constructEventHandlerWithMock(
				ctx,
				ctrl,
				test.cfg,
				test.baseBranch,
				test.resultHasDiff,
				test.resultState,
				test.resultStateDescription)
			assert.Equal(t, test.isError, eventHandler.PullRequestOpened(ctx) != nil)
		})
	}
}
