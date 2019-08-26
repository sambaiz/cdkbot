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

func TestRunner_Deploy(t *testing.T) {
	tests := []struct {
		title                  string
		inUserName             string
		inStacks               []string
		cfg                    config.Config
		baseBranch             string
		labels                 map[string]constant.Label
		resultHasDiff          bool
		retState               *resultState
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
			resultHasDiff:          false,
			retState:            newResultState(constant.StateMergeReady, "No diffs. Let's merge!"),
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
			resultHasDiff:          true,
			retState:            newResultState(constant.StateNeedDeploy, "Fix if needed and complete deploy."),
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
			resultHasDiff:          true,
			retState:            newResultState(constant.StateError,"user sambaiz is not allowed to deploy"),
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

		// hasOutdatedDiffs()
		platformClient.EXPECT().GetPullRequest(ctx).Return(&platform.PullRequest{
			Labels: labels,
		}, nil)

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
			true,
			cfg,
			&platform.PullRequest{
				Number:         1,
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

		openPRs := []platform.PullRequest{
			{
				Number: 1,
				BaseBranch: baseBranch,
			},
			{
				Number: 2,
				BaseBranch: baseBranch,
			},
			{
				Number: 3,
				BaseBranch: "not" + baseBranch,
			},
		}
		platformClient.EXPECT().GetOpenPullRequests(ctx).Return(openPRs, nil)
		cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
		if len(stacks) == 0 {
			stacks = []string{"Stack1", "Stack2"}
			cdkClient.EXPECT().List(cdkPath, target.Contexts).Return(stacks, nil)
		}
		result := "result"
		cdkClient.EXPECT().Deploy(cdkPath, strings.Join(stacks, " "), target.Contexts).Return(result, nil)
		platformClient.EXPECT().AddLabel(ctx, constant.LabelDeployed).Return(nil)
		cdkClient.EXPECT().Diff(cdkPath, "", target.Contexts).Return("", resultHasDiff)
		message := "Success :tada:"
		if resultHasDiff {
			message = "To be continued."
		}
		platformClient.EXPECT().CreateComment(ctx, fmt.Sprintf("### cdk deploy\n```\n%s\n```\n%s", result, message))
		if !resultHasDiff {
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
			title: "base branch is not same",
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
			title: "number is same",
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
