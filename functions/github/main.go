package main

import (
	"context"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/go-github/v26/github"
	"github.com/sambaiz/cdkbot/functions/github/client"
	"github.com/sambaiz/cdkbot/functions/github/event"
	"go.uber.org/zap"
)

var (
	logger        *zap.Logger
	appID         = os.Getenv("GITHUB_APP_ID")
	privateKeyArn = os.Getenv("PRIVATE_KEY_SECRET_ARN")
	webhookSecret = os.Getenv("GITHUB_WEBHOOK_SECRET")
)

func initLogger() {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	logger = zapLogger
}

type response events.APIGatewayProxyResponse

func handler(req events.APIGatewayProxyRequest) (response, error) {
	ctx := context.Background()
	if err := github.ValidateSignature(req.Headers["X-Hub-Signature"], []byte(req.Body), []byte(webhookSecret)); err != nil {
		logger.Info("Signature is invalid", zap.Error(err))
		return response{
			StatusCode: http.StatusBadRequest,
		}, nil
	}
	hook, err := github.ParseWebHook(req.Headers["X-GitHub-Event"], []byte(req.Body))
	if err != nil {
		logger.Error("Failed to parse hook", zap.Error(err))
		return response{
			StatusCode: http.StatusBadRequest,
		}, nil
	}

	cli := client.New(ctx)
	switch hook := hook.(type) {
	case *github.PullRequestEvent:
		err = event.HandlePullRequestEvent(ctx, hook, cli)
	case *github.IssueCommentEvent:
		err = event.HandleIssueCommentEvent(ctx, hook, cli)
	}
	if err != nil {
		logger.Error("Failed to event an event", zap.Error(err))
		return response{
			StatusCode: http.StatusInternalServerError,
		}, err
	}
	return response{
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	initLogger()
	lambda.Start(handler)
}
