package command

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/functions/operation/cdk"
	"github.com/sambaiz/cdkbot/functions/operation/config"
	"github.com/sambaiz/cdkbot/functions/operation/constant"
	"github.com/sambaiz/cdkbot/functions/operation/git"
	"github.com/sambaiz/cdkbot/functions/operation/platform"
)

// Runnerer is interface of Runner
type Runnerer interface {
	Diff(
		ctx context.Context,
		cdkPath string,
		contexts map[string]string,
	) (bool, error)
	Deploy(
		ctx context.Context,
		cdkPath string,
		contexts map[string]string,
		cfg *config.Config,
		userName string,
	) (bool, error)
}

// Runner runs commands
type Runner struct {
	platform platform.Clienter
	git      git.Clienter
	config   config.Readerer
	cdk      cdk.Clienter
}

// NewRunner Runenrn
func NewRunner(client platform.Clienter, cloneURL string) *Runner {
	return &Runner{
		platform: client,
		git:      git.NewClient(cloneURL),
		config:   new(config.Reader),
		cdk:      new(cdk.Client),
	}
}

func (r *Runner) updateStatus(
	ctx context.Context,
	f func() (constant.State, string, error),
) error {
	if err := r.platform.SetStatus(ctx, constant.StateRunning, ""); err != nil {
		return err
	}
	if err := r.platform.AddLabel(ctx, constant.LabelRunning); err != nil {
		return err
	}
	state, statusDescription, err := f()
	defer func() {
		r.platform.SetStatus(
			ctx,
			state,
			statusDescription,
		)
		r.platform.RemoveLabel(ctx, constant.LabelRunning)
	}()
	if err != nil {
		return err
	}
	return nil
}


const clonePath = "/tmp/repo"

func (r *Runner) setup(ctx context.Context) (string, *config.Config, *config.Target, error) {
	hash, err := r.platform.GetPullRequestLatestCommitHash(ctx)
	if err != nil {
		return "", nil, nil, err
	}
	baseBranch, err := r.platform.GetPullRequestBaseBranch(ctx)
	if err != nil {
		return "", nil, nil, err
	}
	if err := r.git.Clone(clonePath, &hash); err != nil {
		return "", nil, nil, err
	}
	if err := r.git.Merge(clonePath, fmt.Sprintf("remotes/origin/%s", baseBranch)); err != nil {
		return "", nil, nil, err
	}

	cfg, err := r.config.Read(fmt.Sprintf("%s/cdkbot.yml", clonePath))
	if err != nil {
		return "", nil, nil, err
	}
	cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
	target, ok := cfg.Targets[baseBranch]
	if !ok {
		return cdkPath, cfg, nil, nil
	}
	if err := r.cdk.Setup(cdkPath); err != nil {
		return "", nil, nil, err
	}
	return cdkPath, cfg, &target, nil
}

// Run a command
func (r *Runner) Run(ctx context.Context, command string, userName string) error {
	if command == "/diff" {
		return r.Diff(ctx)
	} else if command == "/deploy" {
		return r.Deploy(ctx, userName)
	}
	return nil
}
