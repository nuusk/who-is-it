package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"gopkg.in/validator.v1"

	"github.com/google/uuid"
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

// BodyRequest is our self-made struct to process JSON request from Client
type BodyRequest struct {
	ImageBase64 string `json:"image_base64" validate:"nonzero"`
	FileName    string `json:"file_name" validate:"nonzero"`
}

// BodyResponse is our self-made struct to build response for Client
type BodyResponse struct {
	FileName string `json:"file_name"`
	Id       string `json:"id"`
}

func UploadHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// connect to s3
	sess := session.Must(session.NewSession())
	uploader := s3manager.NewUploader(sess)

	uid := uuid.New()

	// Output: 9m4e2mr0ui3e8a215n4g
	// BodyRequest will be used to take the json response from client and build it
	var bodyRequest BodyRequest

	log.Print("request")
	log.Print(request.Body)
	// Unmarshal the json, return 404 if error
	err := json.Unmarshal([]byte(request.Body), &bodyRequest)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusNotFound}, nil
	}
	valid, _ := validator.Validate(bodyRequest)
	log.Print("valid")
	log.Print(valid)
	if !valid {
		return events.APIGatewayProxyResponse{Body: "Validation Error", StatusCode: http.StatusBadRequest}, nil
	}

	log.Print("bodyRequest")
	log.Print(bodyRequest)

	// We will build the BodyResponse and send it back in json form
	bodyResponse := BodyResponse{
		FileName: bodyRequest.FileName + " ok",
		Id:       uid.String(),
	}
	log.Print("bodyResponse")
	log.Print(bodyResponse)

	decoded, err := base64.StdEncoding.DecodeString(bodyRequest.ImageBase64)
	if err != nil {
		log.Printf("error decoding image base 64: %s", err.Error())
		return events.APIGatewayProxyResponse{Body: "error decoding image base 64\n", StatusCode: http.StatusBadRequest}, nil
	}

	bucket := os.Getenv("Bucket")
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(uid.String() + ".png"),
		Body:   bytes.NewReader(decoded),
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		log.Printf("There was an issue uploading to s3: %s", err.Error())
		return events.APIGatewayProxyResponse{Body: "Unable to save response\n", StatusCode: http.StatusBadRequest}, nil
	}

	// Marshal the response into json bytes, if error return 404
	response, err := json.Marshal(&bodyResponse)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 404}, nil
	}

	return events.APIGatewayProxyResponse{Body: string(response), StatusCode: 200}, nil
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
	lambda.Start(UploadHandler)
}
