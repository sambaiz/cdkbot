package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type response events.APIGatewayProxyResponse

func handler(event events.APIGatewayProxyRequest) (response, error) {
	payload, err := json.Marshal(event)
	if err != nil {
		return response{
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	sess := session.New()
	sqsSvc := sqs.New(sess)
	if _, err := sqsSvc.SendMessage(&sqs.SendMessageInput{
		MessageBody:    aws.String(string(payload)),
		QueueUrl:       aws.String(os.Getenv("OPERATION_QUEUE_URL")),
		MessageGroupId: aws.String("group"),
	}); err != nil {
		fmt.Println(err.Error())
		return response{
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	ecsSvc := ecs.New(sess)
	if _, err := ecsSvc.UpdateService(&ecs.UpdateServiceInput{
		Cluster:      aws.String(os.Getenv("TASK_ECS_CLUSTER_ARN")),
		DesiredCount: aws.Int64(1),
		Service:      aws.String(os.Getenv("OPERATION_SERVICE_ARN")),
	}); err != nil {
		fmt.Println(err.Error())
		return response{
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	return response{
		StatusCode: http.StatusOK,
	}, nil
}

func main() {
	lambda.Start(handler)
}
