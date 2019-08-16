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

func TestEventHandlerIssueCommentCreated(t *testing.T) {
	tests := []struct {
		title         string
		inComment     string
		cfg           config.Config
		baseBranch    string
		resultHasDiff bool
		isError       bool
	}{
		{
			title:     "no targets are matched",
			inComment: "/deploy TestStack",
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"master": {},
				},
			},
			baseBranch: "develop",
		},
		{
			title:     "comment diff and has diffs",
			inComment: "/diff",
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
		{
			title:     "comment deploy and has no diffs",
			inComment: "/deploy TestStack",
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
		},
	}

	constructEventHandlerWithMock := func(
		ctx context.Context,
		ctrl *gomock.Controller,
		comment string,
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
		cmd := parseCommand(comment)
		if cmd.action == actionDiff {
			// doActionDiff()
			result := "result"
			cdkClient.EXPECT().Diff(cdkPath, cmd.args, target.Contexts).Return(result, resultHasDiff)
			platformClient.EXPECT().CreateComment(
				ctx,
				fmt.Sprintf("### cdk diff %s\n```\n%s\n```", cmd.args, result),
			).Return(nil)
			platformClient.EXPECT().RemoveLabel(ctx, constant.LabelOutdatedDiff).Return(nil)
		} else if cmd.action == actionDeploy {
			platformClient.EXPECT().AddLabel(ctx, constant.LabelDeploying).Return(nil)

			// doActionDeploy()
			result := "result"
			cdkClient.EXPECT().Deploy(cdkPath, cmd.args, target.Contexts).Return(result, nil)
			cdkClient.EXPECT().Diff(cdkPath, "", target.Contexts).Return("", resultHasDiff)
			platformClient.EXPECT().CreateComment(
				ctx,
				fmt.Sprintf("### cdk deploy %s\n```\n%s\n```\n%s", cmd.args, result, "All stacks have been deployed :tada:"),
			).Return(nil)
			platformClient.EXPECT().AddLabelToOtherPRs(ctx, constant.LabelOutdatedDiff).Return(nil)

			platformClient.EXPECT().RemoveLabel(ctx, constant.LabelDeploying).Return(nil)
		}

		if resultHasDiff {
			platformClient.EXPECT().RemoveLabel(ctx, constant.LabelNoDiff).Return(nil)
		} else {
			platformClient.EXPECT().AddLabel(ctx, constant.LabelNoDiff).Return(nil)
		}

		// updateStatus()
		if resultHasDiff {
			platformClient.EXPECT().SetStatus(ctx, constant.StateNeedDeploy, "Diffs still remain").Return(nil)
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
			eventHandler := constructEventHandlerWithMock(ctx, ctrl, test.inComment, cloneURL, test.cfg, test.baseBranch, test.resultHasDiff)
			assert.Equal(t, test.isError, eventHandler.CommentCreated(ctx, test.inComment) != nil)
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
