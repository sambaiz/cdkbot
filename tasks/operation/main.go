package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/sambaiz/cdkbot/tasks/operation/logger"
	"github.com/sambaiz/cdkbot/tasks/operation/platform/github"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"
)

func main() {
	logger := logger.New()
	flag.Parse()
	arg := flag.Arg(0)
	if arg == "" {
		logger.Error("request must be passed in argument")
		return
	}
	var req events.APIGatewayProxyRequest
	if err := json.Unmarshal([]byte(arg), &req); err != nil {
		logger.Error("error", zap.Error(err))
		return
	}
	ctx := context.Background()
	var err error
	switch os.Getenv("PLATFORM") {
	case "github":
		err = github.Handler(ctx, req, logger)
	default:
		err = fmt.Errorf("unknown platform %s is setted", os.Getenv("PLATFORM"))
	}
	if err != nil {
		logger.Error("error", zap.Error(err))
	}
}
