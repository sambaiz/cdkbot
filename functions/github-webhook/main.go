package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	sdk "github.com/aws/aws-sdk-go/service/lambda"
)

type response events.APIGatewayProxyResponse

func handler(event events.APIGatewayProxyRequest) (response, error) {
	svc := sdk.New(session.New())
	payload, err := json.Marshal(event)
	if err != nil {
		return response{
			StatusCode: http.StatusInternalServerError,
		}, err
	}
	input := &sdk.InvokeInput{
		FunctionName:   aws.String(os.Getenv("INVOKE_FUNTION_ARN")),
		Payload:        payload,
		InvocationType: aws.String("Event"),
	}
	// avoid GitHub Webhook timeout
	if _, err := svc.Invoke(input); err != nil {
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
