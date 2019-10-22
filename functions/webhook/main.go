package main

import (
	"encoding/json"
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
	_, err := json.Marshal(events.APIGatewayProxyRequest{
		Body: event.Body,
		Headers: event.Headers,
	})
	if err != nil {
		return response{
			StatusCode: http.StatusInternalServerError,
		}, err
	}

	svc := ecs.New(session.New())
	
	input := &ecs.UpdateServiceInput{
		Cluster:                       aws.String(os.Getenv("TASK_ECS_CLUSTER_ARN")),
		DesiredCount:                  aws.Int64(1),
		Service:                       aws.String(os.Getenv("OPERATION_SERVICE_ARN")),
	}
	if _, err := svc.UpdateService(input); err != nil {
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
