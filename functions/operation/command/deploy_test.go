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

func TestRunner_Deploy(t *testing.T) {
	type expected struct {
		comment  string
		outState *resultState
		isError  bool
	}
	type test struct {
		title         string
		inUserName    string
		inStacks      []string
		cfg           config.Config
		baseBranch    string
		deployError   error
		resultHasDiff bool
		diffError     error
		expected      expected
	}
	tests := []test{
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
			baseBranch: "develop",
			expected: expected{
				comment:  "### cdk deploy\n```\nresult\n```\n",
				outState: newResultState(constant.StateMergeReady, "No targets are matched"),
				isError:  false,
			},
		},
		{
			title:      "success and has no diffs",
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
			resultHasDiff: false,
			expected: expected{
				comment:  "### cdk deploy\n```\nresult\n```\n",
				outState: newResultState(constant.StateMergeReady, "No diffs. Let's merge!"),
				isError:  false,
			},
		},
		{
			title:      "success_and_has_diffs",
			inUserName: "sambaiz",
			inStacks:   []string{"Stack1", "Stack2"},
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"develop": {},
				},
			},
			baseBranch:    "develop",
			resultHasDiff: true,
			expected: expected{
				comment:  "### cdk deploy\n```\nresult\n```\n",
				outState: newResultState(constant.StateNotMergeReady, "Go ahead with deploy."),
				isError:  false,
			},
		},
		{
			title:      "cdk_deploy_error",
			inUserName: "sambaiz",
			inStacks:   []string{},
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"develop": {},
				},
			},
			baseBranch:  "develop",
			deployError: errors.New("cdk deploy error"),
			expected: expected{
				comment:  "### cdk deploy\n```\nresult\n```\ncdk deploy error",
				outState: newResultState(constant.StateNotMergeReady, "Fix codes"),
				isError:  false,
			},
		},
		{
			title:      "cdk_diff_error",
			inUserName: "sambaiz",
			inStacks:   []string{},
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"develop": {},
				},
			},
			baseBranch: "develop",
			diffError:  errors.New("cdk diff error"),
			expected: expected{
				comment:  "### cdk deploy\n```\nresult\n```\ncdk diff error",
				outState: newResultState(constant.StateNotMergeReady, "Fix codes"),
				isError:  false,
			},
		},
		{
			title:      "the_user_is_not_allowed_to_deploy",
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
				outState: newResultState(constant.StateNotMergeReady, "user sambaiz is not allowed to deploy"),
				isError:  false,
			},
		},
		{
			title:      "other_open_PR_is_been_deploying",
			inUserName: "sambaiz",
			inStacks:   []string{},
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"branch_deploying": {},
				},
			},
			baseBranch: "branch_deploying",
			expected: expected{
				outState: newResultState(constant.StateNotMergeReady, "deployed PR #4 is still opened. First /deploy and merge it, or /rollback."),
				isError:  false,
			},
		},
	}

	constructRunnerWithMock := func(
		ctx context.Context,
		ctrl *gomock.Controller,
		test test,
	) *Runner {
		platformClient := platformMock.NewMockClienter(ctrl)
		gitClient := gitMock.NewMockClienter(ctrl)
		configClient := configMock.NewMockReaderer(ctrl)
		cdkClient := cdkMock.NewMockClienter(ctrl)

		// hasOutdatedDiffs()
		platformClient.EXPECT().GetPullRequest(ctx).Return(&platform.PullRequest{}, nil)

		// updateStatus()
		platformClient.EXPECT().SetStatus(ctx, constant.StateRunning, "").Return(nil)
		platformClient.EXPECT().AddLabel(ctx, constant.LabelRunning).Return(nil)
		platformClient.EXPECT().SetStatus(ctx, test.expected.outState.state, test.expected.outState.description).Return(nil)
		platformClient.EXPECT().RemoveLabel(ctx, constant.LabelRunning).Return(nil)

		constructSetupMock(
			ctx,
			platformClient,
			gitClient,
			configClient,
			cdkClient,
			true,
			test.cfg,
			&platform.PullRequest{
				Number:         1,
				BaseBranch:     test.baseBranch,
				BaseCommitHash: "basehash",
				HeadCommitHash: "headhash",
			},
		)
		target, ok := test.cfg.Targets[test.baseBranch]
		if !ok || !test.cfg.IsUserAllowedDeploy(test.inUserName) {
			return &Runner{
				platform: platformClient,
				git:      gitClient,
				config:   configClient,
				cdk:      cdkClient,
			}
		}

		openPRs := []platform.PullRequest{
			{
				Number:     1,
				BaseBranch: test.baseBranch,
			},
			{
				Number:     2,
				BaseBranch: test.baseBranch,
			},
			{
				Number:     3,
				BaseBranch: "not" + test.baseBranch,
			},
			{
				Number:     4,
				BaseBranch:  "branch_deploying",
				Labels:      map[string]constant.Label{constant.LabelDeployed.Name: constant.LabelDeployed},
			},
		}
		platformClient.EXPECT().GetOpenPullRequests(ctx).Return(openPRs, nil)
		if test.baseBranch == "branch_deploying" {
			return &Runner{
				platform: platformClient,
				git:      gitClient,
				config:   configClient,
				cdk:      cdkClient,
			}
		}

		cdkPath := fmt.Sprintf("%s/%s", clonePath, test.cfg.CDKRoot)
		if len(test.inStacks) == 0 {
			test.inStacks = []string{"Stack1", "Stack2"}
			cdkClient.EXPECT().List(cdkPath, target.Contexts).Return(test.inStacks, nil)
		}
		result := "result"
		cdkClient.EXPECT().Deploy(cdkPath, strings.Join(test.inStacks, " "), target.Contexts).Return(result, test.deployError)
		if test.deployError == nil {
			cdkClient.EXPECT().Diff(cdkPath, "", target.Contexts).Return("", test.resultHasDiff, test.diffError)
		}

		platformClient.EXPECT().AddLabel(ctx, constant.LabelDeployed).Return(nil)
		platformClient.EXPECT().CreateComment(ctx, test.expected.comment)

		if test.deployError != nil || test.diffError != nil {
			return &Runner{
				platform: platformClient,
				git:      gitClient,
				config:   configClient,
				cdk:      cdkClient,
			}
		}

		if !test.resultHasDiff {
			platformClient.EXPECT().MergePullRequest(ctx, "automatically merged by cdkbot").Return(nil)
			// add label to PR with not same number and same base branch
			platformClient.EXPECT().AddLabelToOtherPR(ctx, constant.LabelOutdatedDiff, openPRs[1].Number).Return(nil)
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
			runner := constructRunnerWithMock(ctx, ctrl, test)
			assert.Equal(t, test.expected.isError,
				runner.Deploy(ctx, test.inUserName, test.inStacks) != nil,
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
			title:  "doesn't_have",
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
			platformClient.EXPECT().GetPullRequest(ctx).Return(&platform.PullRequest{
				Labels: test.labels,
			}, nil)
			runner := &Runner{
				platform: platformClient,
			}
			has, err := runner.hasOutdatedDiffLabel(ctx)
			assert.Equal(t, test.out, has)
			assert.Equal(t, test.isError, err != nil)
		})
	}
}

func TestExistsOtherDeployedSameBasePRs(t *testing.T) {
	tests := []struct {
		title      string
		inOtherPRs []platform.PullRequest
		inPR       *platform.PullRequest
		outNumber  int
		outExists  bool
	}{
		{
			title: "exists",
			inOtherPRs: []platform.PullRequest{
				{
					Number:     2,
					BaseBranch: "develop",
					Labels: map[string]constant.Label{
						constant.LabelDeployed.Name: constant.LabelDeployed,
					},
				},
			},
			inPR: &platform.PullRequest{
				Number:     1,
				BaseBranch: "develop",
			},
			outNumber: 2,
			outExists: true,
		},
		{
			title: "base_branch_is_not_same",
			inOtherPRs: []platform.PullRequest{
				{
					Number:     2,
					BaseBranch: "master",
					Labels: map[string]constant.Label{
						constant.LabelDeployed.Name: constant.LabelDeployed,
					},
				},
			},
			inPR: &platform.PullRequest{
				Number:     1,
				BaseBranch: "develop",
			},
			outExists: false,
		},
		{
			title: "number_is_same",
			inOtherPRs: []platform.PullRequest{
				{
					Number:     1,
					BaseBranch: "develop",
					Labels: map[string]constant.Label{
						constant.LabelDeployed.Name: constant.LabelDeployed,
					},
				},
			},
			inPR: &platform.PullRequest{
				Number:     1,
				BaseBranch: "develop",
			},
			outExists: false,
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			number, exists := existsOtherDeployedSameBasePRs(test.inOtherPRs, test.inPR)
			assert.Equal(t, test.outNumber, number)
			assert.Equal(t, test.outExists, exists)
		})
	}
}
