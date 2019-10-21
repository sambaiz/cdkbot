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
)

type response events.APIGatewayProxyResponse

func handler(event events.APIGatewayProxyRequest) (response, error) {
	// Container Overrides length must be at most 8192 so it must be reduced.
	payload, err := json.Marshal(events.APIGatewayProxyRequest{
		Body: event.Body,
		Headers: event.Headers,
	})
	if err != nil {
		return response{
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	svc := ecs.New(session.New())
	input := &ecs.RunTaskInput{
		Cluster:        aws.String(os.Getenv("TASK_ECS_CLUSTER_ARN")),
		LaunchType:     aws.String("FARGATE"),
		NetworkConfiguration: &ecs.NetworkConfiguration{
			AwsvpcConfiguration: &ecs.AwsVpcConfiguration{
				Subnets:        aws.StringSlice([]string{os.Getenv("SUBNET_ID")}),
			},
		},
		TaskDefinition: aws.String(os.Getenv("OPERATION_TASK_DEFINITION_ARN")),
		Count:          aws.Int64(1),
		Overrides: &ecs.TaskOverride{
			ContainerOverrides: []*ecs.ContainerOverride{
				&ecs.ContainerOverride{
					Name:    aws.String("cdkbot-operation"),
					// Command: []*string{aws.String(string(payload))},
				},
			},
		},
	}
	if _, err := svc.RunTask(input); err != nil {
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
