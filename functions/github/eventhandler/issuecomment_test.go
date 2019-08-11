package eventhandler

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sambaiz/cdkbot/functions/github/client"
	githubClientMock "github.com/sambaiz/cdkbot/functions/github/client/mock"
	cdkMock "github.com/sambaiz/cdkbot/lib/cdk/mock"
	"github.com/sambaiz/cdkbot/lib/config"
	configMock "github.com/sambaiz/cdkbot/lib/config/mock"
	gitMock "github.com/sambaiz/cdkbot/lib/git/mock"
	"github.com/stretchr/testify/assert"
)

func TestEventHandlerIssueCommentCreated(t *testing.T) {
	tests := []struct {
		title      string
		in         issueCommentEvent
		cfg        config.Config
		baseBranch string
		isError    bool
	}{
		{
			title: "comment diff and has diff",
			in: issueCommentEvent{
				ownerName:   "owner",
				repoName:    "repo",
				issueNumber: 1,
				commentBody: "/diff",
				cloneURL:    "http://github.com/sambaiz/cdkbot",
			},
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
			isError:    false,
		},
		{
			title: "comment deploy and success",
			in: issueCommentEvent{
				ownerName:   "owner",
				repoName:    "repo",
				issueNumber: 1,
				commentBody: "/deploy TestStack",
				cloneURL:    "http://github.com/sambaiz/cdkbot",
			},
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
			isError:    false,
		},
		{
			title: "no targets are matched",
			in: issueCommentEvent{
				ownerName:   "owner",
				repoName:    "repo",
				issueNumber: 1,
				commentBody: "/deploy TestStack",
				cloneURL:    "http://github.com/sambaiz/cdkbot",
			},
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"master": {
						Contexts: map[string]string{
							"env": "prd",
						},
					},
				},
			},
			baseBranch: "develop",
			isError:    false,
		},
	}

	newEventHandlerWithMock := func(
		ctx context.Context,
		ctrl *gomock.Controller,
		event issueCommentEvent,
		cfg config.Config,
		baseBranch string,
	) *EventHandler {
		githubClient := githubClientMock.NewMockClienter(ctrl)
		gliClient := gitMock.NewMockClienter(ctrl)
		configClient := configMock.NewMockReaderer(ctrl)
		cdkClient := cdkMock.NewMockClienter(ctrl)

		hash := "hash"
		githubClient.EXPECT().CreateStatusOfLatestCommit(
			ctx, event.ownerName, event.repoName, event.issueNumber, client.StatePending).Return(nil)
		githubClient.EXPECT().GetPullRequestLatestCommitHash(
			ctx, event.ownerName, event.repoName, event.issueNumber).Return(hash, nil)
		gliClient.EXPECT().Clone(
			event.cloneURL,
			clonePath,
			&hash,
		).Return(nil)
		configClient.EXPECT().Read(fmt.Sprintf("%s/cdkbot.yml", clonePath)).Return(&cfg, nil)
		githubClient.EXPECT().GetPullRequestBaseBranch(ctx, event.ownerName, event.repoName, event.issueNumber).Return(baseBranch, nil)
		target, ok := cfg.Targets[baseBranch]
		if !ok {
			return &EventHandler{
				cli:    githubClient,
				git:    gliClient,
				config: configClient,
				cdk:    cdkClient,
			}
		}

		cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
		cdkClient.EXPECT().Setup(cdkPath).Return(nil)

		cmd := parseCommand(event.commentBody)
		if cmd.action == actionDiff {
			// doActionDiff()
			result := "result"
			cdkClient.EXPECT().Diff(cdkPath, cmd.args, target.Contexts).Return(result, true)
			githubClient.EXPECT().CreateComment(
				ctx,
				event.ownerName,
				event.repoName,
				event.issueNumber,
				fmt.Sprintf("### cdk diff %s\n```%s```", cmd.args, result),
			).Return(nil)
			githubClient.EXPECT().CreateStatusOfLatestCommit(
				ctx, event.ownerName, event.repoName, event.issueNumber, client.StateFailure).Return(nil)
		} else if cmd.action == actionDeploy {
			// doActionDeploy()
			result := "result"
			cdkClient.EXPECT().Deploy(cdkPath, cmd.args, target.Contexts).Return(result, nil)
			cdkClient.EXPECT().Diff(cdkPath, "", target.Contexts).Return("", false)
			githubClient.EXPECT().CreateComment(
				ctx,
				event.ownerName,
				event.repoName,
				event.issueNumber,
				fmt.Sprintf("### cdk deploy %s\n```%s```\n%s", cmd.args, result, "All stacks have been deployed :tada:"),
			).Return(nil)
			githubClient.EXPECT().CreateStatusOfLatestCommit(
				ctx, event.ownerName, event.repoName, event.issueNumber, client.StateSuccess).Return(nil)
		}

		return &EventHandler{
			cli:    githubClient,
			git:    gliClient,
			config: configClient,
			cdk:    cdkClient,
		}
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			eventHandler := newEventHandlerWithMock(ctx, ctrl, test.in, test.cfg, test.baseBranch)
			assert.Equal(t, test.isError, eventHandler.issueCommentCreated(ctx, test.in) != nil)
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
