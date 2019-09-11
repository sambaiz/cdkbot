package command

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/functions/operation/platform"
	"strings"
	"testing"

	"github.com/sambaiz/cdkbot/functions/operation/constant"

	"errors"
	"github.com/golang/mock/gomock"
	cdkMock "github.com/sambaiz/cdkbot/functions/operation/cdk/mock"
	"github.com/sambaiz/cdkbot/functions/operation/config"
	configMock "github.com/sambaiz/cdkbot/functions/operation/config/mock"
	gitMock "github.com/sambaiz/cdkbot/functions/operation/git/mock"
	platformMock "github.com/sambaiz/cdkbot/functions/operation/platform/mock"
	"github.com/stretchr/testify/assert"
)

func TestRunner_Rollback(t *testing.T) {
	type expected struct {
		comment  string
		outState *resultState
		isError  bool
	}
	tests := []struct {
		title         string
		inUserName    string
		inStacks      []string
		cfg           config.Config
		baseBranch    string
		labels        map[string]constant.Label
		deployError   error
		resultHasDiff bool
		diffError     error
		expected      expected
	}{
		{
			title:      "no_targets_are_matched",
			inUserName: "sambaiz",
			inStacks:   []string{},
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"master": {},
				},
			},
			baseBranch:    "develop",
			labels:        map[string]constant.Label{constant.LabelDeployed.Name: constant.LabelDeployed},
			resultHasDiff: false,
			expected: expected{
				comment: "",
				outState:      newResultState(constant.StateMergeReady, "No targets are matched"),
				isError: false,
			},
		},
		{
			title:      "has_no_diffs",
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
			baseBranch:    "develop",
			labels:        map[string]constant.Label{constant.LabelDeployed.Name: constant.LabelDeployed},
			resultHasDiff: false,
			expected: expected{
				comment: "### cdk deploy (rollback)\n```\nresult\n```\nRollback is completed.",
				outState:     newResultState(constant.StateNotMergeReady, "Run /deploy after reviewed"),
				isError: false,
			},
		},
		{
			title:      "has_diffs",
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
			baseBranch:    "develop",
			labels:        map[string]constant.Label{constant.LabelDeployed.Name: constant.LabelDeployed},
			resultHasDiff: true,
			expected: expected{
				comment: "### cdk deploy (rollback)\n```\nresult\n```\nTo be continued.",
				outState:     newResultState(constant.StateNotMergeReady, "Run /deploy after reviewed"),
				isError: false,
			},
		},
		{
			title:      "cdk_deploy_error",
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
			baseBranch:  "develop",
			labels:      map[string]constant.Label{constant.LabelDeployed.Name: constant.LabelDeployed},
			deployError: errors.New("cdk deploy error"),
			expected: expected{
				comment: "### cdk deploy (rollback)\n```\nresult\n```\ncdk deploy error",
				outState:     newResultState(constant.StateNotMergeReady, "Fix codes"),
				isError: false,
			},
		},
		{
			title:      "cdk_diff_error",
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
			baseBranch: "develop",
			labels:     map[string]constant.Label{constant.LabelDeployed.Name: constant.LabelDeployed},
			diffError:  errors.New("cdk diff error"),
			expected: expected{
				comment: "### cdk deploy (rollback)\n```\nresult\n```\ncdk diff error",
				outState:     newResultState(constant.StateNotMergeReady, "Fix codes"),
				isError: false,
			},
		},
		{
			title:      "user_is_not_allowed_to_deploy",
			inUserName: "sambaiz",
			inStacks:   []string{},
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"develop": {},
				},
				DeployUsers: []string{"foobar"},
			},
			baseBranch:    "develop",
			labels:        map[string]constant.Label{},
			resultHasDiff: true,
			expected: expected{
				comment: "",
				outState:     newResultState(constant.StateNotMergeReady, "user sambaiz is not allowed to deploy"),
				isError: false,
			},
		},
		{
			title:      "PR_is_not_deployed",
			inUserName: "sambaiz",
			inStacks:   []string{},
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"develop": {},
				},
				DeployUsers: []string{"foobar"},
			},
			baseBranch:    "develop",
			resultHasDiff: true,
			expected: expected{
				comment: "",
				outState:     newResultState(constant.StateNotMergeReady, "user sambaiz is not allowed to deploy"),
				isError: false,
			},
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
		deployError error,
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
		cdkClient.EXPECT().Deploy(cdkPath, strings.Join(stacks, " "), target.Contexts).Return(result, deployError)
		if deployError == nil {
			cdkClient.EXPECT().Diff(cdkPath, "", target.Contexts).Return("", resultHasDiff, diffError)
		}
		platformClient.EXPECT().CreateComment(ctx, expected.comment)
		if deployError != nil || diffError != nil {
			return &Runner{
				platform: platformClient,
				git:      gitClient,
				config:   configClient,
				cdk:      cdkClient,
			}
		}
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
				test.deployError,
				test.resultHasDiff,
				test.diffError,
				test.expected)
			assert.Equal(t, test.expected.isError,
				runner.Rollback(ctx, test.inUserName, test.inStacks) != nil,
			)
		})
	}
}
