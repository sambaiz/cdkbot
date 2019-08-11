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

func TestEventHandlerUpdateStatus(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	githubClient := githubClientMock.NewMockClienter(ctrl)
	ownerName := "owner"
	repoName := "repo"
	issueNumber := 1
	resultState := client.StateSuccess
	githubClient.EXPECT().CreateStatusOfLatestCommit(ctx, ownerName, repoName, issueNumber, client.StatePending).Return(nil)
	githubClient.EXPECT().CreateStatusOfLatestCommit(ctx, ownerName, repoName, issueNumber, resultState).Return(nil)
	eventHandler := EventHandler{
		cli: githubClient,
	}
	assert.Nil(t, eventHandler.updateStatus(
		ctx,
		ownerName,
		repoName,
		issueNumber,
		func() (client.State, error) {
			return resultState, nil
		},
	))
}

func TestEventHandlerSetup(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	ownerName := "owner"
	repoName := "repo"
	issueNumber := 1
	cloneURL := "https://github.com/sambaiz/cdkbot"
	baseBranch := "develop"
	cfg := config.Config{
		CDKRoot: ".",
		Targets: map[string]config.Target{
			baseBranch: {},
		},
	}
	githubClient, gitClient, configClient, cdkClient := constructSetupMocks(
		ctx,
		ctrl,
		ownerName,
		repoName,
		issueNumber,
		cloneURL,
		cfg,
		baseBranch,
	)
	client := &EventHandler{
		cli:    githubClient,
		git:    gitClient,
		config: configClient,
		cdk:    cdkClient,
	}
	cdkPath, retCfg, retTarget, err := client.setup(
		ctx,
		ownerName,
		repoName,
		issueNumber,
		cloneURL,
	)
	assert.Equal(t, fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot), cdkPath)
	assert.Equal(t, *retCfg, cfg)
	assert.Equal(t, *retTarget, cfg.Targets[baseBranch])
	assert.Nil(t, err)
}

func constructSetupMocks(
	ctx context.Context,
	ctrl *gomock.Controller,
	ownerName,
	repoName string,
	issueNumber int,
	cloneURL string,
	cfg config.Config,
	baseBranch string,
) (
	*githubClientMock.MockClienter,
	*gitMock.MockClienter,
	*configMock.MockReaderer,
	*cdkMock.MockClienter,
) {
	githubClient := githubClientMock.NewMockClienter(ctrl)
	gitClient := gitMock.NewMockClienter(ctrl)
	configClient := configMock.NewMockReaderer(ctrl)
	cdkClient := cdkMock.NewMockClienter(ctrl)

	hash := "hash"
	githubClient.EXPECT().GetPullRequestLatestCommitHash(
		ctx, ownerName, repoName, issueNumber).Return(hash, nil)
	gitClient.EXPECT().Clone(
		cloneURL,
		clonePath,
		&hash,
	).Return(nil)
	configClient.EXPECT().Read(fmt.Sprintf("%s/cdkbot.yml", clonePath)).Return(&cfg, nil)
	githubClient.EXPECT().GetPullRequestBaseBranch(ctx, ownerName, repoName, issueNumber).Return(baseBranch, nil)
	_, ok := cfg.Targets[baseBranch]
	if !ok {
		return githubClient, gitClient, configClient, cdkClient
	}

	cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
	cdkClient.EXPECT().Setup(cdkPath).Return(nil)

	return githubClient, gitClient, configClient, cdkClient
}
