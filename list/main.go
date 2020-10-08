package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Response events.APIGatewayProxyResponse

type Image struct {
	ID       string `json:"id"`
	FileName string `json:"fileName"`
	Url      string `json:"url"`
}

func ListHandler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// connect to s3
	sess := session.Must(session.NewSession())
	dyna := dynamodb.New(sess)

	table := os.Getenv("Table")
	input := &dynamodb.ScanInput{
		ExpressionAttributeNames: map[string]*string{
			"#ID":       aws.String("ID"),
			"#url":      aws.String("url"),
			"#fileName": aws.String("fileName"),
		},
		// ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
		// 	":a": {
		// 		S: aws.String("No One You Know"),
		// 	},
		// },
		// FilterExpression:     aws.String("Artist = :a"),
		ProjectionExpression: aws.String("#ID, #url, #fileName"),
		TableName:            aws.String(table),
	}

	result, err := dyna.Scan(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				fmt.Println(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Println(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
			case dynamodb.ErrCodeRequestLimitExceeded:
				fmt.Println(dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
			case dynamodb.ErrCodeInternalServerError:
				fmt.Println(dynamodb.ErrCodeInternalServerError, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
	}

	var images []Image
	for _, i := range result.Items {
		image := Image{}

		err = dynamodbattribute.UnmarshalMap(i, &image)
		fmt.Println("unmarshalmap")
		fmt.Println(image)
		images = append(images, image)

		if err != nil {
			fmt.Println("error while unmarshalling:")
			fmt.Println(err.Error())
			return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 404}, nil
		}
	}

	// Marshal the response into json bytes, if error return 404
	response, err := json.Marshal(&images)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 404}, nil
	}
	fmt.Println("response")
	fmt.Println(response)

	return events.APIGatewayProxyResponse{Body: string(response), StatusCode: 200}, nil
}

func main() {
	lambda.Start(ListHandler)
}
