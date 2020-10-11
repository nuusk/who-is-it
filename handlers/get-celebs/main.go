package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pietersweter/who-is-it/pkg/awshelpers"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/rs/zerolog/log"
)

const (
	tableRef = "Table"
)

// Image is used to store images in S3
type Image struct {
	ID       string `json:"id"`
	FileName string `json:"fileName"`
	URL      string `json:"url"`
}

// GetCelebsHandler executes on GET requests on /celeb endpoint
// returns a list of celebrities curently recognized
func GetCelebsHandler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sess := session.Must(session.NewSession())
	dyna := dynamodb.New(sess)

	table := os.Getenv(tableRef)
	input := &dynamodb.ScanInput{
		ExpressionAttributeNames: map[string]*string{
			"#id":       aws.String("id"),
			"#url":      aws.String("url"),
			"#fileName": aws.String("fileName"),
		},
		ProjectionExpression: aws.String("#id, #url, #fileName"),
		TableName:            aws.String(table),
	}

	result, err := dyna.Scan(input)
	if err != nil {
		awshelpers.HandleDynamoDBError(err)
		return events.APIGatewayProxyResponse{Body: "error scanning dynamodb records", StatusCode: http.StatusInternalServerError}, nil
	}

	var images []Image
	for _, i := range result.Items {
		image := Image{}

		err = dynamodbattribute.UnmarshalMap(i, &image)
		if err != nil {
			log.Error().Err(err).Msgf("error while unmarshalling images")
			return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusInternalServerError}, nil
		}
		images = append(images, image)
		log.Error().Err(err).Msgf("new url [%v] appended\nnew list: %v", image, images)
	}

	res, err := json.Marshal(&images)
	if err != nil {
		log.Error().Err(err).Msgf("error while marshalling response")
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 404}, nil
	}
	fmt.Println("res")
	fmt.Println(res)

	return events.APIGatewayProxyResponse{Body: string(res), StatusCode: 200}, nil
}

func main() {
	lambda.Start(GetCelebsHandler)
}
