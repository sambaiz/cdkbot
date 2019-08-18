package main

import (
	"context"
	"fmt"
	"github.com/sambaiz/cdkbot/functions/operation/platform/github"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"go.uber.org/zap"
)

var logger *zap.Logger

func initLogger() {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	logger = zapLogger
}

func handler(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ctx := context.Background()
	switch os.Getenv("PLATFORM") {
	case "github":
		return github.Handler(ctx, req, logger)
	}
	return events.APIGatewayProxyResponse{}, fmt.Errorf("unknown PLATFORM %s is setted", os.Getenv("PLATFORM"))
}

func main() {
	initLogger()
	lambda.Start(handler)
}
