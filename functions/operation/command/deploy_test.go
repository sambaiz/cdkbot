package command

import (
	"context"
	"fmt"
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

func TestRunner_Deploy(t *testing.T) {
	tests := []struct {
		title                  string
		inUserName             string
		cfg                    config.Config
		baseBranch             string
		labels                 map[string]constant.Label
		resultHasDiff          bool
		resultState            constant.State
		resultStateDescription string
		isError                bool
	}{
		{
			title:      "no targets are matched",
			inUserName: "sambaiz",
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
			title:      "has no diffs",
			inUserName: "sambaiz",
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
		{
			title:      "user is not allowed to deploy",
			inUserName: "sambaiz",
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"develop": {},
				},
				DeployUsers: []string{"foobar"},
			},
			baseBranch:             "develop",
			resultHasDiff:          true,
			resultState:            constant.StateError,
			resultStateDescription: "user sambaiz is not allowed to deploy",
		},
	}

	constructRunnerWithMock := func(
		ctx context.Context,
		ctrl *gomock.Controller,
		userName string,
		cfg config.Config,
		baseBranch string,
		labels map[string]constant.Label,
		resultHasDiff bool,
		resultState constant.State,
		resultStateDescription string,
	) *Runner {
		platformClient := platformMock.NewMockClienter(ctrl)
		gitClient := gitMock.NewMockClienter(ctrl)
		configClient := configMock.NewMockReaderer(ctrl)
		cdkClient := cdkMock.NewMockClienter(ctrl)

		// hasOutdatedDiffs()
		platformClient.EXPECT().GetPullRequestLabels(ctx).Return(labels, nil)

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
		if !cfg.IsUserAllowedDeploy(userName) {
			return &Runner{
				platform: platformClient,
				git:      gitClient,
				config:   configClient,
				cdk:      cdkClient,
			}
		}

		platformClient.EXPECT().AddLabelToOtherPRs(ctx, constant.LabelOutdatedDiff).Return(nil)
		cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
		stacks := []string{"Stack1", "Stack2"}
		cdkClient.EXPECT().List(cdkPath, target.Contexts).Return(stacks, nil)
		result := "result"
		cdkClient.EXPECT().Deploy(cdkPath, strings.Join(stacks, " "), target.Contexts).Return(result, nil)
		cdkClient.EXPECT().Diff(cdkPath, "", target.Contexts).Return("", resultHasDiff)
		message := "All stacks have been deployed :tada:"
		if resultHasDiff {
			message = "Some stacks are failed to deploy... Don't give up!"
		}
		platformClient.EXPECT().CreateComment(ctx, fmt.Sprintf("### cdk deploy\n```\n%s\n```\n%s", result, message))
		if !resultHasDiff {
			platformClient.EXPECT().MergePullRequest(ctx, "automatically merged by cdkbot").Return(nil)
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
				test.cfg,
				test.baseBranch,
				test.labels,
				test.resultHasDiff,
				test.resultState,
				test.resultStateDescription)
			assert.Equal(t, test.isError,
				runner.Deploy(ctx, test.inUserName) != nil,
			)
		})
	}
}

func TestRunner_hasOutdatedDiff(t *testing.T) {
	tests := []struct {
		title   string
		labels  map[string]constant.Label
		out     bool
		isError bool
	}{
		{
			title:  "has",
			labels: map[string]constant.Label{constant.LabelOutdatedDiff.Name: constant.LabelOutdatedDiff},
			out:    true,
		},
		{
			title:  "doesn't have",
			labels: map[string]constant.Label{},
			out:    false,
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			platformClient := platformMock.NewMockClienter(ctrl)
			platformClient.EXPECT().GetPullRequestLabels(ctx).Return(test.labels, nil)
			runner := &Runner{
				platform: platformClient,
			}
			has, err := runner.hasOutdatedDiffLabel(ctx)
			assert.Equal(t, test.out, has)
			assert.Equal(t, test.isError, err != nil)
		})
	}
}
