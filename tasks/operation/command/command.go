package command

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/tasks/operation/cdk"
	"github.com/sambaiz/cdkbot/tasks/operation/config"
	"github.com/sambaiz/cdkbot/tasks/operation/constant"
	"github.com/sambaiz/cdkbot/tasks/operation/git"
	"github.com/sambaiz/cdkbot/tasks/operation/logger"
	"github.com/sambaiz/cdkbot/tasks/operation/platform"
	"go.uber.org/zap"
	"os/exec"
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
	logger   logger.Loggerer
}

// NewRunner creates Runner
func NewRunner(client platform.Clienter, cloneURL string, logger logger.Loggerer) *Runner {
	return &Runner{
		platform: client,
		git:      git.NewClient(cloneURL),
		config:   new(config.Reader),
		cdk:      new(cdk.Client),
		logger:   logger,
	}
}

type resultState struct {
	state       constant.State
	description string
}

func newResultState(state constant.State, description string) *resultState {
	return &resultState{
		state:       state,
		description: description,
	}
}

func (r *Runner) updateStatus(
	ctx context.Context,
	f func() (*resultState, error),
) error {
	if err := r.platform.SetStatus(ctx, constant.StateRunning, ""); err != nil {
		return err
	}
	if err := r.platform.AddLabel(ctx, constant.LabelRunning); err != nil {
		return err
	}
	state, err := f()
	if err := r.platform.RemoveLabel(ctx, constant.LabelRunning); err != nil {
		r.logger.Error("remove label error", zap.Error(err))
	}
	if err != nil {
		if err := r.platform.SetStatus(
			ctx,
			constant.StateError,
			err.Error(),
		); err != nil {
			r.logger.Error("set status error", zap.Error(err))
		}
		return err
	}
	if err := r.platform.SetStatus(
		ctx,
		state.state,
		state.description,
	); err != nil {
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
		if err := r.git.Checkout(clonePath, pr.BaseBranch); err != nil {
			return "", nil, nil, nil, err
		}
		if err := r.git.Merge(clonePath, pr.HeadCommitHash); err != nil {
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

	// override cdkbot.yml & cdk.yml of base branch
	if err := r.git.CheckoutFile(clonePath, "cdkbot.yml", pr.BaseBranch); err != nil {
		return "", nil, nil, nil, err
	}
	if err := r.git.CheckoutFile(fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot), "cdk.json", pr.BaseBranch); err != nil {
		return "", nil, nil, nil, err
	}

	if err := r.cdk.Setup(cdkPath); err != nil {
		return "", nil, nil, nil, err
	}

	for _, preCommand := range cfg.PreCommands {
		command := strings.Split(preCommand, " ")
		cmd := exec.Command(command[0], command[1:]...)
		cmd.Dir = cdkPath
		if out, err := cmd.CombinedOutput(); err != nil || cmd.ProcessState.ExitCode() != 0 {
			return "", nil, nil, nil, fmt.Errorf("preCommand %s failed: %s %v", preCommand, string(out), err)
		}
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
