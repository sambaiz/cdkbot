package command

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/functions/operation/platform"
	"strings"
	"testing"

	"github.com/sambaiz/cdkbot/functions/operation/constant"

	"github.com/golang/mock/gomock"
	cdkMock "github.com/sambaiz/cdkbot/functions/operation/cdk/mock"
	"github.com/sambaiz/cdkbot/functions/operation/config"
	configMock "github.com/sambaiz/cdkbot/functions/operation/config/mock"
	gitMock "github.com/sambaiz/cdkbot/functions/operation/git/mock"
	platformMock "github.com/sambaiz/cdkbot/functions/operation/platform/mock"
	"github.com/stretchr/testify/assert"
)

func TestRunner_Rollback(t *testing.T) {
	tests := []struct {
		title                  string
		inUserName             string
		inStacks               []string
		cfg                    config.Config
		baseBranch             string
		labels                 map[string]constant.Label
		resultHasDiff          bool
		retState            *resultState
		isError                bool
	}{
		{
			title:      "no targets are matched",
			inUserName: "sambaiz",
			inStacks:   []string{},
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"master": {},
				},
			},
			baseBranch:             "develop",
			labels:                 map[string]constant.Label{ constant.LabelDeployed.Name: constant.LabelDeployed },
			resultHasDiff:          false,
			retState:            newResultState(constant.StateMergeReady, "No targets are matched"),
		},
		{
			title:      "has no diffs",
			inUserName: "sambaiz",
			inStacks:   []string{},
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
			labels:                 map[string]constant.Label{ constant.LabelDeployed.Name: constant.LabelDeployed },
			resultHasDiff:          false,
			retState:            newResultState(constant.StateNeedDeploy, "Run /deploy after reviewed"),
		},
		{
			title:      "has diffs",
			inUserName: "sambaiz",
			inStacks:   []string{"Stack1", "Stack2"},
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
			labels:                 map[string]constant.Label{ constant.LabelDeployed.Name: constant.LabelDeployed },
			resultHasDiff:          true,
			retState:            newResultState(constant.StateNeedDeploy, "Run /deploy after reviewed"),
		},
		{
			title:      "user is not allowed to deploy",
			inUserName: "sambaiz",
			inStacks:   []string{},
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"develop": {},
				},
				DeployUsers: []string{"foobar"},
			},
			baseBranch:             "develop",
			labels:                 map[string]constant.Label{},
			resultHasDiff:          true,
			retState:            newResultState(constant.StateError, "user sambaiz is not allowed to deploy"),
		},
		{
			title:      "PR is not deployed",
			inUserName: "sambaiz",
			inStacks:   []string{},
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"develop": {},
				},
				DeployUsers: []string{"foobar"},
			},
			baseBranch:             "develop",
			resultHasDiff:          true,
			retState:            newResultState(constant.StateError, "user sambaiz is not allowed to deploy"),
		},
	}

	constructRunnerWithMock := func(
		ctx context.Context,
		ctrl *gomock.Controller,
		userName string,
		stacks []string,
		cfg config.Config,
		baseBranch string,
		labels map[string]constant.Label,
		resultHasDiff bool,
		retState *resultState,
	) *Runner {
		platformClient := platformMock.NewMockClienter(ctrl)
		gitClient := gitMock.NewMockClienter(ctrl)
		configClient := configMock.NewMockReaderer(ctrl)
		cdkClient := cdkMock.NewMockClienter(ctrl)

		// updateStatus()
		platformClient.EXPECT().SetStatus(ctx, constant.StateRunning, "").Return(nil)
		platformClient.EXPECT().AddLabel(ctx, constant.LabelRunning).Return(nil)
		platformClient.EXPECT().SetStatus(ctx, retState.state, retState.description).Return(nil)
		platformClient.EXPECT().RemoveLabel(ctx, constant.LabelRunning).Return(nil)

		constructSetupMock(
			ctx,
			platformClient,
			gitClient,
			configClient,
			cdkClient,
			false,
			cfg,
			&platform.PullRequest{
				BaseBranch:     baseBranch,
				BaseCommitHash: "basehash",
				HeadCommitHash: "headhash",
				Labels:         labels,
			},
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
		if !cfg.IsUserAllowedDeploy(userName) {
			return &Runner{
				platform: platformClient,
				git:      gitClient,
				config:   configClient,
				cdk:      cdkClient,
			}
		}

		cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
		if len(stacks) == 0 {
			stacks = []string{"Stack1", "Stack2"}
			cdkClient.EXPECT().List(cdkPath, target.Contexts).Return(stacks, nil)
		}
		result := "result"
		cdkClient.EXPECT().Deploy(cdkPath, strings.Join(stacks, " "), target.Contexts).Return(result, nil)
		cdkClient.EXPECT().Diff(cdkPath, "", target.Contexts).Return("", resultHasDiff)
		message := "Rollback is completed."
		if resultHasDiff {
			message = "To be continued."
		}
		platformClient.EXPECT().CreateComment(ctx, fmt.Sprintf("### cdk deploy (rollback)\n```\n%s\n```\n%s", result, message))
		if !resultHasDiff {
			platformClient.EXPECT().RemoveLabel(ctx, constant.LabelDeployed).Return(nil)
		}

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
				test.inUserName,
				test.inStacks,
				test.cfg,
				test.baseBranch,
				test.labels,
				test.resultHasDiff,
				test.retState)
			assert.Equal(t, test.isError,
				runner.Rollback(ctx, test.inUserName, test.inStacks) != nil,
			)
		})
	}
}
