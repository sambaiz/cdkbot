package eventhandler

import (
	"context"
	"fmt"
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

func TestEventHandlerUpdateStatus(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	platformClient := platformMock.NewMockClienter(ctrl)
	resultState := constant.StateMergeReady
	statusDescription := "description"
	platformClient.EXPECT().SetStatus(ctx, constant.StateRunning, "").Return(nil)
	platformClient.EXPECT().AddLabel(ctx, constant.LabelRunning).Return(nil)
	platformClient.EXPECT().SetStatus(ctx, resultState, statusDescription).Return(nil)
	platformClient.EXPECT().RemoveLabel(ctx, constant.LabelRunning).Return(nil)
	eventHandler := EventHandler{
		platform: platformClient,
	}
	assert.Nil(t, eventHandler.updateStatus(
		ctx,
		func() (constant.State, string, error) {
			return resultState, statusDescription, nil
		},
	))
}

func TestEventHandlerSetup(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	platformClient := platformMock.NewMockClienter(ctrl)
	gitClient := gitMock.NewMockClienter(ctrl)
	configClient := configMock.NewMockReaderer(ctrl)
	cdkClient := cdkMock.NewMockClienter(ctrl)

	baseBranch := "develop"
	cfg := config.Config{
		CDKRoot: ".",
		Targets: map[string]config.Target{
			baseBranch: {},
		},
	}
	constructSetupMock(
		ctx,
		platformClient,
		gitClient,
		configClient,
		cdkClient,
		cfg,
		baseBranch,
	)
	eventHandler := &EventHandler{
		platform: platformClient,
		git:      gitClient,
		config:   configClient,
		cdk:      cdkClient,
	}
	cdkPath, retCfg, retTarget, err := eventHandler.setup(ctx)
	assert.Equal(t, fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot), cdkPath)
	assert.Equal(t, *retCfg, cfg)
	assert.Equal(t, *retTarget, cfg.Targets[baseBranch])
	assert.Nil(t, err)
}

func constructSetupMock(
	ctx context.Context,
	platformClient *platformMock.MockClienter,
	gitClient *gitMock.MockClienter,
	configClient *configMock.MockReaderer,
	cdkClient *cdkMock.MockClienter,
	cfg config.Config,
	baseBranch string,
) {
	hash := "hash"
	platformClient.EXPECT().GetPullRequestLatestCommitHash(ctx).Return(hash, nil)
	platformClient.EXPECT().GetPullRequestBaseBranch(ctx).Return(baseBranch, nil)
	gitClient.EXPECT().Clone(clonePath, &hash).Return(nil)
	gitClient.EXPECT().Merge(clonePath, fmt.Sprintf("remotes/origin/%s", baseBranch)).Return(nil)
	configClient.EXPECT().Read(fmt.Sprintf("%s/cdkbot.yml", clonePath)).Return(&cfg, nil)
	_, ok := cfg.Targets[baseBranch]
	if !ok {
		return
	}

	cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
	cdkClient.EXPECT().Setup(cdkPath).Return(nil)

	return
}
