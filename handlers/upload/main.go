package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rs/zerolog/log"
	"gopkg.in/validator.v1"

	"github.com/google/uuid"
)

// UploadImageRequest is our self-made struct to process JSON request from Client
type UploadImageRequest struct {
	ImageBase64 string `json:"imageBase64" validate:"nonzero"`
	FileName    string `json:"fileName" validate:"nonzero"`
}

// BodyResponse is our self-made struct to build response for Client
type BodyResponse struct {
	FileName string `json:"fileName"`
	ID       string `json:"id"`
}

func getImagePublicURL(key string) string {
	bucketName := "helloworld-dev-storage-1qp3fqg74aaqi"
	region := "eu-west-1"
	url := fmt.Sprintf(
		"http://%s.s3-%s.amazonaws.com/%s",
		bucketName,
		region,
		key,
	)
	return url
}

func getImageNameWithExtension(key string) string {
	name := fmt.Sprintf(
		"%s.png",
		key,
	)
	return name
}

// UploadHandler ...
func UploadHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sess := session.Must(session.NewSession())
	dyna := dynamodb.New(sess)
	uploader := s3manager.NewUploader(sess)

	uid := uuid.New()

	var uploadImageRequest UploadImageRequest

	log.Info().Msgf("got request: %v", request)

	err := json.Unmarshal([]byte(request.Body), &uploadImageRequest)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusNotFound}, nil
	}
	valid, _ := validator.Validate(uploadImageRequest)
	log.Print("valid")
	log.Print(valid)
	if !valid {
		return events.APIGatewayProxyResponse{Body: "Validation Error", StatusCode: http.StatusBadRequest}, nil
	}

	log.Print("uploadImageRequest")
	log.Print(uploadImageRequest)

	// We will build the BodyResponse and send it back in json form
	bodyResponse := BodyResponse{
		FileName: uploadImageRequest.FileName,
		ID:       uid.String(),
	}

	log.Info().Msgf("created bodyResponse: %v", bodyResponse)

	decoded, err := base64.StdEncoding.DecodeString(uploadImageRequest.ImageBase64)
	if err != nil {
		log.Printf("error decoding image base 64: %s", err.Error())
		return events.APIGatewayProxyResponse{Body: "error decoding image base 64\n", StatusCode: http.StatusBadRequest}, nil
	}

	bucket := os.Getenv("Bucket")
	iKey := uid.String()
	iName := getImageNameWithExtension(iKey)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(iName),
		Body:   bytes.NewReader(decoded),
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		log.Printf("There was an issue uploading to s3: %s", err.Error())
		return events.APIGatewayProxyResponse{Body: "Unable to save response\n", StatusCode: http.StatusBadRequest}, nil
	}

	table := os.Getenv("Table")
	newItem := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(iKey),
			},
			"fileName": {
				S: aws.String(uploadImageRequest.FileName),
			},
			"url": {
				S: aws.String(getImagePublicURL(iName)),
			},
		},
		TableName: aws.String(table),
	}
	result, err := dyna.PutItem(newItem)
	if err != nil {
		aerr, _ := err.(awserr.Error)
		if aerr != nil {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				fmt.Println(dynamodb.ErrCodeConditionalCheckFailedException, aerr.Error())
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
			case dynamodb.ErrCodeItemCollectionSizeLimitExceededException:
				fmt.Println(dynamodb.ErrCodeItemCollectionSizeLimitExceededException, aerr.Error())
			case dynamodb.ErrCodeTransactionConflictException:
				fmt.Println(dynamodb.ErrCodeTransactionConflictException, aerr.Error())
			case dynamodb.ErrCodeRequestLimitExceeded:
				fmt.Println(dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
			case dynamodb.ErrCodeInternalServerError:
				fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
			return events.APIGatewayProxyResponse{Body: "error with dyna put item\n", StatusCode: http.StatusBadRequest}, nil
		} else {
			fmt.Println(err.Error())
		}
	}

	fmt.Println(result)

	// Marshal the response into json bytes, if error return 404
	response, err := json.Marshal(&bodyResponse)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 404}, nil
	}

	return events.APIGatewayProxyResponse{Body: string(response), StatusCode: 200}, nil
}

func main() {
	lambda.Start(UploadHandler)
}
