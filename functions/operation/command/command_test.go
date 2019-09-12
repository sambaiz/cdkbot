package command

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/functions/operation/logger"
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

func TestRunner_updateStatus(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	platformClient := platformMock.NewMockClienter(ctrl)
	retState := newResultState(constant.StateMergeReady, "description")
	platformClient.EXPECT().SetStatus(ctx, constant.StateRunning, "").Return(nil)
	platformClient.EXPECT().AddLabel(ctx, constant.LabelRunning).Return(nil)
	platformClient.EXPECT().SetStatus(ctx, retState.state, retState.description).Return(nil)
	platformClient.EXPECT().RemoveLabel(ctx, constant.LabelRunning).Return(nil)
	runner := Runner{
		platform: platformClient,
	}
	assert.Nil(t, runner.updateStatus(
		ctx,
		func() (*resultState, error) {
			return retState, nil
		},
	))
}

func TestRunner_setup(t *testing.T) {
	for _, cloneHead := range []bool{true, false} {
		t.Run(fmt.Sprintf("cloneHead: %v", cloneHead), func(t *testing.T) {
			ctx := context.Background()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
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
			pr := &platform.PullRequest{
				BaseBranch:     baseBranch,
				BaseCommitHash: "basehash",
				HeadCommitHash: "headhash",
				Labels:         nil,
			}
			constructSetupMock(
				ctx,
				platformClient,
				gitClient,
				configClient,
				cdkClient,
				cloneHead,
				cfg,
				pr,
			)
			runner := &Runner{
				platform: platformClient,
				git:      gitClient,
				config:   configClient,
				cdk:      cdkClient,
				logger:   logger.MockLogger{},
			}
			cdkPath, retCfg, retTarget, outpr, err := runner.setup(ctx, cloneHead)
			assert.Equal(t, fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot), cdkPath)
			assert.Equal(t, *retCfg, cfg)
			assert.Equal(t, *retTarget, cfg.Targets[baseBranch])
			assert.Equal(t, pr, outpr)
			assert.Nil(t, err)
		})
	}
}

func constructSetupMock(
	ctx context.Context,
	platformClient *platformMock.MockClienter,
	gitClient *gitMock.MockClienter,
	configClient *configMock.MockReaderer,
	cdkClient *cdkMock.MockClienter,
	cloneHead bool,
	cfg config.Config,
	pr *platform.PullRequest,
) {
	platformClient.EXPECT().GetPullRequest(ctx).Return(pr, nil)
	if cloneHead {
		gitClient.EXPECT().Clone(clonePath, &pr.HeadCommitHash).Return(nil)
		gitClient.EXPECT().Merge(clonePath, fmt.Sprintf("remotes/origin/%s", pr.BaseBranch)).Return(nil)
	} else {
		gitClient.EXPECT().Clone(clonePath, &pr.BaseCommitHash).Return(nil)
	}
	configClient.EXPECT().Read(fmt.Sprintf("%s/cdkbot.yml", clonePath)).Return(&cfg, nil)
	_, ok := cfg.Targets[pr.BaseBranch]
	if !ok {
		return
	}

	gitClient.EXPECT().Checkout(clonePath, "cdkbot.yml", pr.BaseBranch).Return(nil)
	gitClient.EXPECT().Checkout(fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot), "cdk.json", pr.BaseBranch).Return(nil)

	cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
	cdkClient.EXPECT().Setup(cdkPath).Return(nil)

	return
}

func TestParseStacks(t *testing.T) {
	tests := []struct {
		title   string
		in      string
		out     []string
		isError bool
	}{
		{
			title: "success",
			in:    "/deploy Stack1 Stack2",
			out:   []string{"Stack1", "Stack2"},
		},
		{
			title:   "inavlid stackname",
			in:      "/deploy Stack1 $tack2",
			isError: true,
		},
	}
	for _, test := range tests {
		stacks, err := parseStacks(test.in)
		assert.Equal(t, test.out, stacks)
		assert.Equal(t, test.isError, err != nil)
	}
}

func TestValidateStackName(t *testing.T) {
	tests := []struct {
		title   string
		in      string
		isError bool
	}{
		{
			title:   "valid",
			in:      "Stack-1",
			isError: false,
		},
		{
			title:   "invalid_character",
			in:      "Sta`ck1",
			isError: true,
		},
		{
			title:   "too_long",
			in:      strings.Repeat("A", 129),
			isError: true,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.isError, validateStackName(test.in) != nil)
	}
}
