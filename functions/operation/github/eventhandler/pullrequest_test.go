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
		title                string
		in                   pullRequestEvent
		cfg                  config.Config
		baseBranch           string
		resultHasDiff        bool
		outState             client.State
		outStatusDescription string
		isError              bool
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
			baseBranch:           "develop",
			outState:             client.StateSuccess,
			outStatusDescription: "No targets are matched",
		},
		{
			title: "target is matched and has diffs",
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
			baseBranch:           "develop",
			resultHasDiff:        true,
			outState:             client.StateFailure,
			outStatusDescription: "There are diffs",
		},
	}

	constructEventHandlerWithMock := func(
		ctx context.Context,
		ctrl *gomock.Controller,
		event pullRequestEvent,
		cfg config.Config,
		baseBranch string,
		resultHasDiff bool,
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
		cdkClient.EXPECT().Diff(cdkPath, "", target.Contexts).Return(result, resultHasDiff)
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
			eventHandler := constructEventHandlerWithMock(ctx, ctrl, test.in, test.cfg, test.baseBranch, test.resultHasDiff)
			state, statusDescription, err := eventHandler.pullRequestOpened(ctx, test.in)
			assert.Equal(t, test.isError, err != nil)
			assert.Equal(t, test.outState, state)
			assert.Equal(t, test.outStatusDescription, statusDescription)
		})
	}
}
