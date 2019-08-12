package eventhandler

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sambaiz/cdkbot/functions/operation/github/client"
	"github.com/sambaiz/cdkbot/lib/config"
	"github.com/stretchr/testify/assert"
)

func TestEventHandlerPullRequestOpened(t *testing.T) {
	tests := []struct {
		title      string
		in         pullRequestEvent
		cfg        config.Config
		baseBranch string
		out        client.State
		isError    bool
	}{
		{
			title: "no targets are matched",
			in: pullRequestEvent{
				ownerName: "owner",
				repoName:  "repo",
				prNumber:  1,
				cloneURL:  "http://github.com/sambaiz/cdkbot",
			},
			cfg: config.Config{
				CDKRoot: ".",
				Targets: map[string]config.Target{
					"master": {},
				},
			},
			baseBranch: "develop",
			out:        client.StateSuccess,
		},
		{
			title: "target is matched",
			in: pullRequestEvent{
				ownerName: "owner",
				repoName:  "repo",
				prNumber:  1,
				cloneURL:  "http://github.com/sambaiz/cdkbot",
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
			out:        client.StateFailure,
		},
	}

	constructEventHandlerWithMock := func(
		ctx context.Context,
		ctrl *gomock.Controller,
		event pullRequestEvent,
		cfg config.Config,
		baseBranch string,
	) *EventHandler {
		githubClient, gitClient, configClient, cdkClient := constructSetupMocks(
			ctx,
			ctrl,
			event.ownerName,
			event.repoName,
			event.prNumber,
			event.cloneURL,
			cfg,
			baseBranch,
		)

		if _, ok := cfg.Targets[baseBranch]; !ok {
			return &EventHandler{
				cli:    githubClient,
				git:    gitClient,
				config: configClient,
				cdk:    cdkClient,
			}
		}

		target := cfg.Targets[baseBranch]
		cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
		result := "result"
		cdkClient.EXPECT().Diff(cdkPath, "", target.Contexts).Return(result, true)
		githubClient.EXPECT().CreateComment(
			ctx,
			event.ownerName,
			event.repoName,
			event.prNumber,
			fmt.Sprintf("### cdk diff\n```%s```\n", result),
		).Return(nil)

		return &EventHandler{
			cli:    githubClient,
			git:    gitClient,
			config: configClient,
			cdk:    cdkClient,
		}
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			eventHandler := constructEventHandlerWithMock(ctx, ctrl, test.in, test.cfg, test.baseBranch)
			state, err := eventHandler.pullRequestOpened(ctx, test.in)
			assert.Equal(t, test.isError, err != nil)
			assert.Equal(t, test.out, state)
		})
	}
}
