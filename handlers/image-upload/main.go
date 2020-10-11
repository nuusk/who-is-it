package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/pietersweter/who-is-it/pkg/awshelpers"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/rs/zerolog/log"
	"gopkg.in/validator.v1"

	"github.com/google/uuid"
)

const (
	tableRef  = "Table"
	bucketRef = "Bucket"
	regionRef = "Region"
)

// ImageUploadRequest is a JSON representation of raw request from the client
type ImageUploadRequest struct {
	ImageBase64 string `json:"imageBase64" validate:"nonzero"`
	FileName    string `json:"fileName" validate:"nonzero"`
	Extension   string `json:"extension" validate:"nonzero,regexp=^(jpg|png)$"`
}

// BodyResponse is used to geneerate JSON response for the client
type BodyResponse struct {
	URL string `json:"url"`
}

// UploadHandler is executed for POST to the /celeb endpoint
// it uploads an image to s3 and returns a public URL to it
func UploadHandler(ctx context.Context, reqRaw events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sess := session.Must(session.NewSession())
	dyna := dynamodb.New(sess)
	uploader := s3manager.NewUploader(sess)

	var reqJSON ImageUploadRequest
	err := json.Unmarshal([]byte(reqRaw.Body), &reqJSON)
	if err != nil {
		log.Error().Err(err).Msgf("error unmarshalling request")
		return events.APIGatewayProxyResponse{Body: "error unmarshalling request", StatusCode: http.StatusBadRequest}, nil
	}

	valid, errs := validator.Validate(reqJSON)
	if !valid {
		for _, err := range errs {
			for _, errMsg := range err {
				log.Error().Err(errMsg).Msgf("validation request failure: %v", errMsg)
			}
		}
		return events.APIGatewayProxyResponse{Body: "validation request body failure\n", StatusCode: http.StatusBadRequest}, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(reqJSON.ImageBase64)
	if err != nil {
		log.Error().Err(err).Msgf("error decoding image base 64: %s", err.Error())
		return events.APIGatewayProxyResponse{Body: "error decoding image base 64\n", StatusCode: http.StatusInternalServerError}, nil
	}

	bucket := os.Getenv(bucketRef)
	uid := uuid.New().String()
	key := awshelpers.GetImageNameWithExtension(uid, reqJSON.Extension)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(decoded),
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		log.Error().Err(err).Msgf("issue uploading to s3: %s", err.Error())
		return events.APIGatewayProxyResponse{Body: "unable to upload to s3\n", StatusCode: http.StatusInternalServerError}, nil
	}
	log.Info().Msgf("uploaded image to s3 with key: %s", key)

	table := os.Getenv(tableRef)
	newItem := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(uid),
			},
			"fileName": {
				S: aws.String(reqJSON.FileName),
			},
			"url": {
				S: aws.String(awshelpers.GetPublicURLFromKey(key)),
			},
		},
		TableName: aws.String(table),
	}

	_, err = dyna.PutItem(newItem)
	if err != nil {
		awshelpers.HandleDynamoDBError(err)
		return events.APIGatewayProxyResponse{Body: "unable to upload to dynamodb\n", StatusCode: http.StatusInternalServerError}, nil
	}
	log.Info().Msgf("uploaded image to dynamodb with key: %s", uid)

	res := BodyResponse{
		URL: awshelpers.GetPublicURLFromKey(key),
	}
	resRaw, err := json.Marshal(&res)
	if err != nil {
		log.Error().Err(err).Msgf("error marshalling response")
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusInternalServerError}, nil
	}

	return events.APIGatewayProxyResponse{Body: string(resRaw), StatusCode: http.StatusOK}, nil
}

func main() {
	lambda.Start(UploadHandler)
}
