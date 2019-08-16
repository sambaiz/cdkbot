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
		title         string
		cfg           config.Config
		baseBranch    string
		resultHasDiff bool
		isError       bool
	}{
		{
			title: "no targets are matched",
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"master": {},
				},
			},
			baseBranch: "develop",
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
			baseBranch:    "develop",
			resultHasDiff: true,
		},
	}

	constructEventHandlerWithMock := func(
		ctx context.Context,
		ctrl *gomock.Controller,
		cloneURL string,
		cfg config.Config,
		baseBranch string,
		resultHasDiff bool,
	) *EventHandler {
		platformClient := platformMock.NewMockClienter(ctrl)
		gitClient := gitMock.NewMockClienter(ctrl)
		configClient := configMock.NewMockReaderer(ctrl)
		cdkClient := cdkMock.NewMockClienter(ctrl)

		platformClient.EXPECT().SetStatus(ctx, constant.StateRunning, "").Return(nil)
		constructSetupMock(
			ctx,
			platformClient,
			gitClient,
			configClient,
			cdkClient,
			cloneURL,
			cfg,
			baseBranch,
		)

		if _, ok := cfg.Targets[baseBranch]; !ok {
			platformClient.EXPECT().SetStatus(ctx, constant.StateMergeReady, "No targets are matched").Return(nil)
			return &EventHandler{
				platform: platformClient,
				git:      gitClient,
				config:   configClient,
				cdk:      cdkClient,
			}
		}

		target, ok := cfg.Targets[baseBranch]
		if !ok {
			platformClient.EXPECT().SetStatus(ctx, constant.StateMergeReady, "No targets are matched").Return(nil)
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

		// updateStatus()
		if resultHasDiff {
			platformClient.EXPECT().SetStatus(ctx, constant.StateNeedDeploy, "There are differences").Return(nil)
		} else {
			platformClient.EXPECT().SetStatus(ctx, constant.StateMergeReady, "No diffs. Let's merge!").Return(nil)
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
			cloneURL := "https://example.com"
			eventHandler := constructEventHandlerWithMock(ctx, ctrl, cloneURL, test.cfg, test.baseBranch, test.resultHasDiff)
			assert.Equal(t, test.isError, eventHandler.PullRequestOpened(ctx) != nil)
		})
	}
}
