package eventhandler

import (
	"context"
	"fmt"

	"github.com/sambaiz/cdkbot/functions/operation/github/client"
	"github.com/sambaiz/cdkbot/lib/cdk"
	"github.com/sambaiz/cdkbot/lib/config"
	"github.com/sambaiz/cdkbot/lib/git"
)

// EventHandler is github events handler
type EventHandler struct {
	cli    client.Clienter
	git    git.Clienter
	config config.Readerer
	cdk    cdk.Clienter
}

// New event handler
func New(ctx context.Context) *EventHandler {
	return &EventHandler{
		cli:    client.New(ctx),
		git:    new(git.Client),
		config: new(config.Reader),
		cdk:    new(cdk.Client),
	}
}

func (e *EventHandler) updateStatus(
	ctx context.Context,
	ownerName string,
	repoName string,
	issueNumber int,
	f func() (client.State, string, error),
) error {
	if err := e.cli.CreateStatusOfLatestCommit(
		ctx,
		ownerName,
		repoName,
		issueNumber,
		client.StatePending,
		nil,
	); err != nil {
		return err
	}
	state, statusDescription, err := f()
	defer func() {
		e.cli.CreateStatusOfLatestCommit(
			ctx,
			ownerName,
			repoName,
			issueNumber,
			state,
			&statusDescription,
		)
	}()
	if err != nil {
		return err
	}
	return nil
}

const clonePath = "/tmp/repo"

func (e *EventHandler) setup(
	ctx context.Context,
	ownerName string,
	repoName string,
	issueNumber int,
	cloneURL string,
) (string, *config.Config, *config.Target, error) {
	hash, err := e.cli.GetPullRequestLatestCommitHash(
		ctx,
		ownerName,
		repoName,
		issueNumber,
	)
	if err != nil {
		return "", nil, nil, err
	}
	if err := e.git.Clone(cloneURL, clonePath, &hash); err != nil {
		return "", nil, nil, err
	}

	cfg, err := e.config.Read(fmt.Sprintf("%s/cdkbot.yml", clonePath))
	if err != nil {
		return "", nil, nil, err
	}
	cdkPath := fmt.Sprintf("%s/%s", clonePath, cfg.CDKRoot)
	baseBranch, err := e.cli.GetPullRequestBaseBranch(
		ctx,
		ownerName,
		repoName,
		issueNumber,
	)
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
