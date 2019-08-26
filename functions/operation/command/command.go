package command

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/functions/operation/cdk"
	"github.com/sambaiz/cdkbot/functions/operation/config"
	"github.com/sambaiz/cdkbot/functions/operation/constant"
	"github.com/sambaiz/cdkbot/functions/operation/git"
	"github.com/sambaiz/cdkbot/functions/operation/platform"
	"regexp"
	"strings"
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

func (r *Runner) setup(ctx context.Context, cloneHead bool) (string, *config.Config, *config.Target, *platform.PullRequest, error) {
	pr, err := r.platform.GetPullRequest(ctx)
	if err != nil {
		return "", nil, nil, nil, err
	}
	if cloneHead {
		if err := r.git.Clone(clonePath, &pr.HeadCommitHash); err != nil {
			return "", nil, nil, nil, err
		}
		if err := r.git.Merge(clonePath, fmt.Sprintf("remotes/origin/%s", pr.BaseBranch)); err != nil {
			return "", nil, nil, nil, err
		}
	} else {
		if err := r.git.Clone(clonePath, &pr.BaseCommitHash); err != nil {
			return "", nil, nil, nil, err
		}
	}

	cfg, err := r.config.Read(fmt.Sprintf("%s/cdkbot.yml", clonePath))
	if err != nil {
		return "", nil, nil, nil, err
	}
	cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
	target, ok := cfg.Targets[pr.BaseBranch]
	if !ok {
		return cdkPath, cfg, nil, nil, nil
	}
	if err := r.cdk.Setup(cdkPath); err != nil {
		return "", nil, nil, nil, err
	}
	return cdkPath, cfg, &target, pr, nil
}

// Run a command
func (r *Runner) Run(ctx context.Context, command string, userName string) error {
	if command == "/diff" {
		return r.Diff(ctx)
	} else if strings.HasPrefix(command, "/deploy") {
		stacks, err := parseStacks(command)
		if err != nil {
			return err
		}
		return r.Deploy(ctx, userName, stacks)
	} else if strings.HasPrefix(command, "/rollback") {
		stacks, err := parseStacks(command)
		if err != nil {
			return err
		}
		return r.Rollback(ctx, userName, stacks)
	}

	return nil
}

func parseStacks(command string) ([]string, error) {
	args := strings.Split(command, " ")
	stacks := []string{}
	if len(args) != 0 {
		stacks = args[1:]
		for _, stack := range stacks {
			if err := validateStackName(stack); err != nil {
				return nil, err
			}
		}
	}
	return stacks, nil
}

// It must start with an alphabetic character and can't be longer than 128 characters.
// A stack name can contain only alphanumeric characters (case-sensitive) and hyphens.
var validStackNameFormat = regexp.MustCompile(`^[a-zA-Z][a-zA-Z\d\-]{0,127}$`)

// Check the stack name so that the command does not contain illegal characters.
func validateStackName(name string) error {
	if !validStackNameFormat.Match([]byte(name)) {
		return fmt.Errorf("Invalid stack name %s", name)
	}
	return nil
}
