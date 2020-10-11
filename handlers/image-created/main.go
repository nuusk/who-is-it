package main

import (
	"context"
	"encoding/json"
	"github.com/pietersweter/who-is-it/pkg/awshelpers"
	"os"

	"github.com/aws/aws-lambda-go/events"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/rs/zerolog/log"
)

// CreatedHandler is triggered after an object is uploaded to s3
// it runs a rekognition call to determine celebrities in the photo
func CreatedHandler(ctx context.Context, event events.SQSEvent) (bool, error) {
	sess := session.Must(session.NewSession())
	svc := rekognition.New(sess)
	dyna := dynamodb.New(sess)

	var s3Ev events.S3Event

	for _, record := range event.Records {
		log.Info().Msgf("%v", record.Body)
		err := json.Unmarshal([]byte(record.Body), &s3Ev)
		if err != nil {
			log.Error().Err(err).Msgf("error unmarshalling sqs event to s3 event")
			return false, err
		}

		for _, s3Record := range s3Ev.Records {

			celebIn := &rekognition.RecognizeCelebritiesInput{
				Image: &rekognition.Image{
					S3Object: &rekognition.S3Object{
						Bucket: aws.String(s3Record.S3.Bucket.Name),
						Name:   aws.String(s3Record.S3.Object.Key),
					},
				},
			}

			celebRes, err := svc.RecognizeCelebrities(celebIn)
			if err != nil {
				awshelpers.HandleRekognitionError(err)
				return false, err
			}

			table := os.Getenv("Table")
			for _, celeb := range celebRes.CelebrityFaces {
				log.Info().Msgf("%s found", *celeb.Name)

				newImageURL := &dynamodb.AttributeValue{
					S: aws.String(awshelpers.GetPublicURLFromRecord(s3Record)),
				}

				var images []*dynamodb.AttributeValue
				images = append(images, newImageURL)

				updated := &dynamodb.UpdateItemInput{
					Key: map[string]*dynamodb.AttributeValue{
						"ID": {
							S: aws.String(*celeb.Id),
						},
					},
					ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
						":images": {
							L: images,
						},
						":empty_list": {
							L: []*dynamodb.AttributeValue{},
						},
					},
					ReturnValues:     aws.String("ALL_NEW"),
					UpdateExpression: aws.String("SET images = list_append(if_not_exists(images, :empty_list), :images)"),
					TableName:        aws.String(table),
				}
				_, err = dyna.UpdateItem(updated)
				if err != nil {
					awshelpers.HandleDynamoDBError(err)
					return false, err
				}
				log.Info().Msgf("%s updated, key: %s", *celeb.Name, s3Record.S3.Object.Key)
			}

			nbNoCelebs := len(celebRes.UnrecognizedFaces)
			log.Info().Msgf("%v unrecognized people", nbNoCelebs)
		}
	}

	return true, nil
}

func main() {
	lambda.Start(CreatedHandler)
}
