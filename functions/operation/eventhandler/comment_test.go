package eventhandler

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

func TestEventHandlerIssueCommentCreated(t *testing.T) {
	tests := []struct {
		title                  string
		inUserName             string
		inComment              string
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
			inComment:  "/deploy",
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
			title:      "comment diff and has diffs",
			inUserName: "sambaiz",
			inComment:  "/diff",
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
			resultStateDescription: "Diffs still remain",
		},
		{
			title:      "comment deploy and has no diffs",
			inUserName: "sambaiz",
			inComment:  "/deploy",
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
			title:      "comment deploy but user is not allowed to deploy",
			inUserName: "sambaiz",
			inComment:  "/deploy",
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
			resultStateDescription: "User sambaiz is not allowed to deploy",
		},
		{
			title:      "comment deploy but since differences are outdated so run /diff instead",
			inUserName: "sambaiz",
			inComment:  "/deploy",
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"develop": {},
				},
				DeployUsers: []string{"foobar"},
			},
			baseBranch:             "develop",
			labels:                 map[string]constant.Label{constant.LabelOutdatedDiff.Name: constant.LabelOutdatedDiff},
			resultHasDiff:          true,
			resultState:            constant.StateNeedDeploy,
			resultStateDescription: "Diffs still remain",
		},
	}

	constructEventHandlerWithMock := func(
		ctx context.Context,
		ctrl *gomock.Controller,
		userName string,
		comment string,
		cfg config.Config,
		baseBranch string,
		labels map[string]constant.Label,
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
			return &EventHandler{
				platform: platformClient,
				git:      gitClient,
				config:   configClient,
				cdk:      cdkClient,
			}
		}

		cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
		cmd := parseCommand(comment)

		if cmd.action == actionDeploy {
			// hasOutdatedDiffLabel()
			platformClient.EXPECT().GetPullRequestLabels(ctx).Return(labels, nil)
			if _, ok := labels[constant.LabelOutdatedDiff.Name]; ok {
				platformClient.EXPECT().CreateComment(ctx, "Differences are outdated. Run /diff instead.").Return(nil)
				cmd.action = actionDiff
				cmd.args = ""
			}
		}

		if cmd.action == actionDiff {
			// doActionDiff()
			previousComments := []platform.Comment{{ID: 1, Body: "### cdk diff\n```\nresult\n```"}}
			platformClient.EXPECT().ListComments(ctx).Return(previousComments, nil)
			result := "result"
			cdkClient.EXPECT().Diff(cdkPath, "", target.Contexts).Return(result, resultHasDiff)
			platformClient.EXPECT().CreateComment(
				ctx,
				fmt.Sprintf("### cdk diff\n```\n%s\n```", result),
			).Return(nil)
			platformClient.EXPECT().DeleteComment(ctx, previousComments[0].ID).Return(nil)
			platformClient.EXPECT().RemoveLabel(ctx, constant.LabelOutdatedDiff).Return(nil)
		} else if cmd.action == actionDeploy {
			if !cfg.IsUserAllowedDeploy(userName) {
				return &EventHandler{
					platform: platformClient,
					git:      gitClient,
					config:   configClient,
					cdk:      cdkClient,
				}
			}

			// doActionDeploy()
			platformClient.EXPECT().AddLabelToOtherPRs(ctx, constant.LabelOutdatedDiff).Return(nil)
			stacks := []string{"Stack1 Stack2"}
			cdkClient.EXPECT().List(cdkPath, target.Contexts).Return(stacks, nil)
			result := "result"
			cdkClient.EXPECT().Deploy(cdkPath, strings.Join(stacks, " "), target.Contexts).Return(result, nil)
			cdkClient.EXPECT().Diff(cdkPath, "", target.Contexts).Return("", resultHasDiff)
			platformClient.EXPECT().CreateComment(
				ctx,
				fmt.Sprintf("### cdk deploy\n```\n%s\n```\n%s", result, "All stacks have been deployed :tada:"),
			).Return(nil)
			if !resultHasDiff {
				platformClient.EXPECT().MergePullRequest(ctx, "automatically merged by cdkbot").Return(nil)
			}
		}

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
				test.inUserName,
				test.inComment,
				test.cfg,
				test.baseBranch,
				test.labels,
				test.resultHasDiff,
				test.resultState,
				test.resultStateDescription)
			assert.Equal(t, test.isError,
				eventHandler.CommentCreated(ctx, test.inUserName, test.inComment) != nil,
			)
		})
	}
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		title   string
		in      string
		out     *command
		isError bool
	}{
		{
			title: "diff",
			in:    "/diff aaa bbb",
			out: &command{
				action: actionDiff,
				args:   "aaa bbb",
			},
		},
		{
			title: "deploy",
			in:    "/deploy aaa bbb",
			out: &command{
				action: actionDeploy,
				args:   "aaa bbb",
			},
		},
		{
			title: "unknown",
			in:    "/unknown aaa bbb",
			out:   nil,
		},
	}
	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			cmd := parseCommand(test.in)
			assert.Equal(t, test.out, cmd)
		})
	}
}

func TestEventHandlerHasOutdatedDiff(t *testing.T) {
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
			labels: map[string]constant.Label{constant.LabelNoDiff.Name: constant.LabelNoDiff},
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
			eventHandler := &EventHandler{
				platform: platformClient,
			}
			has, err := eventHandler.hasOutdatedDiffLabel(ctx)
			assert.Equal(t, test.out, has)
			assert.Equal(t, test.isError, err != nil)
		})
	}
}

func TestEventHandlerDeleteDiffCommentsUpToPreviousDeploy(t *testing.T) {
	diffComment := "### cdk diff\n```\nresult\n```"
	deployComment := "### cdk deploy\n```\nresult\n```"
	tests := []struct {
		title   string
		in      []platform.Comment
		expectedDeletedIDs []int64
		isError bool
	}{
		{
			title:  "Delete diff comments up to previous deploy",
			in: []platform.Comment{
				{
					ID: 1,
					Body: diffComment,
				},
				{
					ID: 2,
					Body: diffComment,
				},
			},
			expectedDeletedIDs: []int64{1, 2},
		},
		{
			title:  "Don't delete anything other than diff comments",
			in: []platform.Comment{
				{
					ID: 1,
					Body: "a",
				},
				{
					ID: 2,
					Body: diffComment,
				},
			},
			expectedDeletedIDs: []int64{2},
		},
		{
			title:  "Don't delete diff comments before previous deploy",
			in: []platform.Comment{
				{
					ID: 1,
					Body: diffComment,
				},
				{
					ID: 2,
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
			eventHandler := &EventHandler{
				platform: platformClient,
			}
			err := eventHandler.deleteDiffCommentsUpToPreviousDeploy(ctx, test.in)
			assert.Equal(t, test.isError, err != nil)
		})
	}
}