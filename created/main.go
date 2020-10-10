package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/rs/zerolog/log"
)

type Celebrity struct {
	Name 	string
	ID 		string
}

// CreatedHandler ...
func CreatedHandler(ctx context.Context, event events.S3Event) (bool, error) {
	sess := session.Must(session.NewSession())
	svc := rekognition.New(sess)
	dyna := dynamodb.New(sess)

	for _, record := range event.Records {
		log.Info().Msgf("%v\n %v", record.S3.Bucket.Name, record.S3.Object.Key)

		celebIn := &rekognition.RecognizeCelebritiesInput{
			Image: &rekognition.Image{
				S3Object: &rekognition.S3Object{
					Bucket: aws.String(record.S3.Bucket.Name),
					Name:   aws.String(record.S3.Object.Key),
				},
			},
		}

		celebRes, err := svc.RecognizeCelebrities(celebIn)
		if err != nil {
			handleRekognitionError(err)
		}

		table := os.Getenv("Table")
		for _, celeb := range celebRes.CelebrityFaces {
			log.Info().Msgf("%s found", *celeb.Name)

			// uid := uuid.New()

			// newItem := &dynamodb.PutItemInput{
			// 	Item: map[string]*dynamodb.AttributeValue{
			// 		"ID": {
			// 			S: aws.String(uid.String()),
			// 		},
			// 		"celebrity_ID": {
			// 			S: aws.String(*celeb.Id),
			// 		},
			// 		"celebrity_name": {
			// 			S: aws.String(*celeb.Name),
			// 		},
			// 		"celebrity_image_url": {
			// 			S: aws.String(record.S3.Bucket.Name),
			// 		},
			// 	},
			// 	TableName: aws.String(table),
			// }
			// _, err := dyna.PutItem(newItem)
			// if err != nil {
			// 	handleDynamoDBError(err)
			// }
			// log.Info().Msgf("%s uploaded into dyna, key: ", *celeb.Name, record.S3.Object.Key)

			newImageURL := &dynamodb.AttributeValue{  
				// S: aws.String(imagePublicURL),
				S: aws.String(getImagePublicURL(record)),
			} 

			var images []*dynamodb.AttributeValue 
			images = append(images, newImageURL)  

			updated := &dynamodb.UpdateItemInput{
				Key: map[string]*dynamodb.AttributeValue{
					"ID": {
						S: aws.String(*celeb.Id),
					},
				},
				ExpressionAttributeValues: map[string]*dynamodb.AttributeValue {
					":images": {  
						L: images,   
						},   
					":empty_list": {   
						L: []*dynamodb.AttributeValue{},  
					},  
				},
				ReturnValues:  aws.String("ALL_NEW"), 
				UpdateExpression: aws.String("SET images = list_append(if_not_exists(images, :empty_list), :images)"), 
				TableName: aws.String(table),
			}
			_, err = dyna.UpdateItem(updated)
			if err != nil {
				handleDynamoDBError(err)
			}
			log.Info().Msgf("%s updated, key: ", *celeb.Name, record.S3.Object.Key)
		}


		nbNoCelebs := len(celebRes.UnrecognizedFaces)
		log.Info().Msgf("%v unrecognized people", nbNoCelebs)

	}

	return true, nil
}

func main() {
	lambda.Start(CreatedHandler)
}

func handleDynamoDBError(err error) {
	aerr, ok := err.(awserr.Error)
	if ok {
		switch aerr.Code() {
		case dynamodb.ErrCodeConditionalCheckFailedException:
			log.Error().Err(err).Msgf(dynamodb.ErrCodeConditionalCheckFailedException, aerr.Error())
		case dynamodb.ErrCodeProvisionedThroughputExceededException:
			log.Error().Err(err).Msgf(dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
		case dynamodb.ErrCodeResourceNotFoundException:
			log.Error().Err(err).Msgf(dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
		case dynamodb.ErrCodeItemCollectionSizeLimitExceededException:
			log.Error().Err(err).Msgf(dynamodb.ErrCodeItemCollectionSizeLimitExceededException, aerr.Error())
		case dynamodb.ErrCodeTransactionConflictException:
			log.Error().Err(err).Msgf(dynamodb.ErrCodeTransactionConflictException, aerr.Error())
		case dynamodb.ErrCodeRequestLimitExceeded:
			log.Error().Err(err).Msgf(dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
		case dynamodb.ErrCodeInternalServerError:
			log.Error().Err(err).Msgf(dynamodb.ErrCodeInternalServerError, aerr.Error())
		default:
			log.Error().Err(err).Msgf(aerr.Error())
		}
		// return events.APIGatewayProxyResponse{Body: "error with dyna put item\n", StatusCode: http.StatusBadRequest}, nil
	} else {
		log.Error().Err(err).Msgf(err.Error())
	}
}

func handleRekognitionError(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case rekognition.ErrCodeInvalidS3ObjectException:
			log.Error().Err(err).Msgf(rekognition.ErrCodeInvalidS3ObjectException, aerr.Error())
		case rekognition.ErrCodeInvalidParameterException:
			log.Error().Err(err).Msgf(rekognition.ErrCodeInvalidParameterException, aerr.Error())
		case rekognition.ErrCodeImageTooLargeException:
			log.Error().Err(err).Msgf(rekognition.ErrCodeImageTooLargeException, aerr.Error())
		case rekognition.ErrCodeAccessDeniedException:
			log.Error().Err(err).Msgf(rekognition.ErrCodeAccessDeniedException, aerr.Error())
		case rekognition.ErrCodeInternalServerError:
			log.Error().Err(err).Msgf(rekognition.ErrCodeInternalServerError, aerr.Error())
		case rekognition.ErrCodeThrottlingException:
			log.Error().Err(err).Msgf(rekognition.ErrCodeThrottlingException, aerr.Error())
		case rekognition.ErrCodeProvisionedThroughputExceededException:
			log.Error().Err(err).Msgf(rekognition.ErrCodeProvisionedThroughputExceededException, aerr.Error())
		case rekognition.ErrCodeInvalidImageFormatException:
			log.Error().Err(err).Msgf(rekognition.ErrCodeInvalidImageFormatException, aerr.Error())
		default:
			log.Error().Err(err).Msgf(aerr.Error())
		}
	} else {
		log.Error().Err(err).Msgf(err.Error())
	}
}

func getImagePublicURL(record events.S3EventRecord) string {
	imagePublicURL := fmt.Sprintf("%s.s3-%s.amazonaws.com/%s",
		record.S3.Bucket.Name, 
		record.AWSRegion, 
	 	record.S3.Object.Key,
	)
	return imagePublicURL
}
