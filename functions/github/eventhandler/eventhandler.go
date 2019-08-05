package eventhandler

import (
	"context"

	"github.com/sambaiz/cdkbot/functions/github/client"
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

func New(ctx context.Context) *EventHandler {
	return &EventHandler{
		cli:    client.New(ctx),
		git:    new(git.Client),
		config: new(config.Reader),
		cdk:    new(cdk.Client),
	}
}
