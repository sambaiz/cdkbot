package eventhandler

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/functions/operation/constant"
	"github.com/sambaiz/cdkbot/functions/operation/platform"

	"github.com/sambaiz/cdkbot/functions/operation/cdk"
	"github.com/sambaiz/cdkbot/functions/operation/config"
	"github.com/sambaiz/cdkbot/functions/operation/git"
	goGit "gopkg.in/src-d/go-git.v4"
)

// EventHandler handles events
type EventHandler struct {
	platform platform.Clienter
	git      git.Clienter
	config   config.Readerer
	cdk      cdk.Clienter
}

// New EventHandler
func New(ctx context.Context, client platform.Clienter, cloneOptions *goGit.CloneOptions) *EventHandler {
	return &EventHandler{
		platform: client,
		git:      git.NewClient(cloneOptions),
		config:   new(config.Reader),
		cdk:      new(cdk.Client),
	}
}

func (e *EventHandler) updateStatus(
	ctx context.Context,
	f func() (constant.State, string, error),
) error {
	if err := e.platform.SetStatus(ctx, constant.StateRunning, ""); err != nil {
		return err
	}
	if err := e.platform.AddLabel(ctx, constant.LabelRunning); err != nil {
		return err
	}
	state, statusDescription, err := f()
	defer func() {
		e.platform.SetStatus(
			ctx,
			state,
			statusDescription,
		)
		e.platform.RemoveLabel(ctx, constant.LabelRunning)
	}()
	if err != nil {
		return err
	}
	return nil
}

const clonePath = "/tmp/repo"

func (e *EventHandler) setup(ctx context.Context) (string, *config.Config, *config.Target, error) {
	hash, err := e.platform.GetPullRequestLatestCommitHash(ctx)
	if err != nil {
		return "", nil, nil, err
	}
	if err := e.git.Clone(clonePath, &hash); err != nil {
		return "", nil, nil, err
	}

	cfg, err := e.config.Read(fmt.Sprintf("%s/cdkbot.yml", clonePath))
	if err != nil {
		return "", nil, nil, err
	}
	cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
	baseBranch, err := e.platform.GetPullRequestBaseBranch(ctx)
	if err != nil {
		return "", nil, nil, err
	}
	target, ok := cfg.Targets[baseBranch]
	if !ok {
		return cdkPath, cfg, nil, nil
	}
	if err := e.cdk.Setup(cdkPath); err != nil {
		return "", nil, nil, err
	}
	return cdkPath, cfg, &target, nil
}
