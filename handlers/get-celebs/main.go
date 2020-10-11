package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/pietersweter/who-is-it/pkg/awshelpers"

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

// Celeb is used to store Celebr name and photos associated with him/her
type Celeb struct {
	Name   string   `json:"celeb_name"`
	Images []string `json:"celeb_images"`
}

// GetCelebsResponse contains array of celebrities with their photos
type GetCelebsResponse struct {
	Celebs []Celeb `json:"celebs"`
}

// GetCelebsHandler executes on GET requests on /celeb endpoint
// returns a list of celebrities curently recognized
func GetCelebsHandler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	sess := session.Must(session.NewSession())
	dyna := dynamodb.New(sess)

	table := os.Getenv(tableRef)
	input := &dynamodb.ScanInput{
		ExpressionAttributeNames: map[string]*string{
			"#celeb_images": aws.String("celeb_images"),
			"#celeb_name":   aws.String("celeb_name"),
		},
		ProjectionExpression: aws.String("#celeb_images, #celeb_name"),
		TableName:            aws.String(table),
	}

	result, err := dyna.Scan(input)
	if err != nil {
		awshelpers.HandleDynamoDBError(err)
		return events.APIGatewayProxyResponse{Body: "error scanning dynamodb records", StatusCode: http.StatusInternalServerError}, nil
	}

	resRaw := GetCelebsResponse{}
	for _, i := range result.Items {
		celeb := Celeb{}

		err = dynamodbattribute.UnmarshalMap(i, &celeb)
		if err != nil {
			log.Error().Err(err).Msgf("error while unmarshalling celebrities")
			return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: http.StatusInternalServerError}, nil
		}

		resRaw.Celebs = append(resRaw.Celebs, celeb)
	}

	res, err := json.Marshal(&resRaw)
	if err != nil {
		log.Error().Err(err).Msgf("error while marshalling response")
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 404}, nil
	}

	return events.APIGatewayProxyResponse{Body: string(res), StatusCode: 200}, nil
}

func main() {
	lambda.Start(GetCelebsHandler)
}
