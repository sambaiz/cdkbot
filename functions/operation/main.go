package main

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/functions/operation/logger"
	"github.com/sambaiz/cdkbot/functions/operation/platform/github"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.uber.org/zap"
)

func handler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	ctx := context.Background()
	logger := logger.New()
	var (
		resp *events.APIGatewayProxyResponse
		err  error
	)
	switch os.Getenv("PLATFORM") {
	case "github":
		resp, err = github.Handler(ctx, req, logger)
	default:
		resp = &events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}
		err = fmt.Errorf("unknown platform %s is setted", os.Getenv("PLATFORM"))
	}
	if err != nil {
		logger.Error("error", zap.Error(err))
	}
	return resp, nil // suppress to retry function
}

func main() {
	lambda.Start(handler)
}
