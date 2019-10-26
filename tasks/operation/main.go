package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/sambaiz/cdkbot/tasks/operation/logger"
	"github.com/sambaiz/cdkbot/tasks/operation/platform/github"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"
)

func main() {
	logger := logger.New()
	sess := session.New()
	sqsSvc := sqs.New(sess)
	for {
		res, err := sqsSvc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(os.Getenv("OPERATION_QUEUE_URL")),
			MaxNumberOfMessages: aws.Int64(1),
		})
		if err != nil {
			logger.Error("receive message error", zap.Error(err))
			break
		}

		if len(res.Messages) == 0 {
			break
		}

		msg := res.Messages[0]
		if _, err := sqsSvc.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      aws.String(os.Getenv("OPERATION_QUEUE_URL")),
			ReceiptHandle: msg.ReceiptHandle,
		}); err != nil {
			logger.Error("delete message error", zap.Error(err))
			break
		}
		var req events.APIGatewayProxyRequest
		if err := json.Unmarshal([]byte(*msg.Body), &req); err != nil {
			logger.Error("unmarshal error", zap.Error(err))
			continue
		}

		ctx := context.Background()
		switch os.Getenv("PLATFORM") {
		case "github":
			err = github.Handler(ctx, req, logger)
		default:
			err = fmt.Errorf("unknown platform %s is setted", os.Getenv("PLATFORM"))
		}
		if err != nil {
			logger.Error("operation error", zap.Error(err))
		}
	}

	ecsSvc := ecs.New(sess)
	if _, err := ecsSvc.UpdateService(&ecs.UpdateServiceInput{
		Cluster:      aws.String(os.Getenv("TASK_ECS_CLUSTER_ARN")),
		DesiredCount: aws.Int64(0),
		Service:      aws.String(os.Getenv("OPERATION_SERVICE_ARN")),
	}); err != nil {
		logger.Error("shutdown task error", zap.Error(err))
	}
}
