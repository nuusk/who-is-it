package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
//
// https://serverless.com/framework/docs/providers/aws/events/apigateway/#lambda-proxy-integration
type Response events.APIGatewayProxyResponse

var invokeCount = 0
var myObjects []*s3.Object

func init() {
	svc := s3.New(session.New())
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String("examplebucket"),
	}
	result, _ := svc.ListObjectsV2(input)
	myObjects = result.Contents
}

func LambdaHandler(ctx context.Context) (Response, error) {
	invokeCount = invokeCount + 1

	resp := Response{
		StatusCode:      200,
		IsBase64Encoded: false,
		Body:            fmt.Sprint(invokeCount),
		Headers: map[string]string{
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "hello-handler",
		},
	}

	log.Print(myObjects)
	return resp, nil
}

// Handler is our lambda handler invoked by the `lambda.Start` function call
// func Handler(ctx context.Context) (Response, error) {
// 	var buf bytes.Buffer

// 	body, err := json.Marshal(map[string]interface{}{
// 		"message": "Go Serverless v1.0! Your function executed successfully!",
// 	})
// 	if err != nil {
// 		return Response{StatusCode: 404}, err
// 	}
// 	json.HTMLEscape(&buf, body)

// 	resp := Response{
// 		StatusCode:      200,
// 		IsBase64Encoded: false,
// 		Body:            buf.String(),
// 		Headers: map[string]string{
// 			"Content-Type":           "application/json",
// 			"X-MyCompany-Func-Reply": "hello-handler",
// 		},
// 	}

// 	return resp, nil
// }

func main() {
	lambda.Start(LambdaHandler)
}
