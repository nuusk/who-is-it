// Package awshelpers provides helpers and handlers for communication and interacting with s3, dynamoDB, rekognition
package awshelpers

import (
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/rs/zerolog/log"
)

const (
	tableRef  = "Table"
	bucketRef = "Bucket"
	regionRef = "Region"
)

// HandleDynamoDBError loops through all possible errors while interacting with dynamodb
// and outputs a desired log
func HandleDynamoDBError(err error) {
	aerr, ok := err.(awserr.Error)
	if ok {
		switch aerr.Code() {
		case dynamodb.ErrCodeConditionalCheckFailedException:
			log.Error().Err(err).Msgf("%v, %v", dynamodb.ErrCodeConditionalCheckFailedException, aerr.Error())
		case dynamodb.ErrCodeProvisionedThroughputExceededException:
			log.Error().Err(err).Msgf("%v, %v", dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
		case dynamodb.ErrCodeResourceNotFoundException:
			log.Error().Err(err).Msgf("%v, %v", dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
		case dynamodb.ErrCodeItemCollectionSizeLimitExceededException:
			log.Error().Err(err).Msgf("%v, %v", dynamodb.ErrCodeItemCollectionSizeLimitExceededException, aerr.Error())
		case dynamodb.ErrCodeTransactionConflictException:
			log.Error().Err(err).Msgf("%v, %v", dynamodb.ErrCodeTransactionConflictException, aerr.Error())
		case dynamodb.ErrCodeRequestLimitExceeded:
			log.Error().Err(err).Msgf("%v, %v", dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
		case dynamodb.ErrCodeInternalServerError:
			log.Error().Err(err).Msgf("%v, %v", dynamodb.ErrCodeInternalServerError, aerr.Error())
		default:
			log.Error().Err(err).Msgf(aerr.Error())
		}
	} else {
		log.Error().Err(err).Msgf(err.Error())
	}
}

// HandleRekognitionError loops through all possible errors while interacting with rekognition api
// and outputs a desired log
func HandleRekognitionError(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case rekognition.ErrCodeInvalidS3ObjectException:
			log.Error().Err(err).Msgf("%v, %v", rekognition.ErrCodeInvalidS3ObjectException, aerr.Error())
		case rekognition.ErrCodeInvalidParameterException:
			log.Error().Err(err).Msgf("%v, %v", rekognition.ErrCodeInvalidParameterException, aerr.Error())
		case rekognition.ErrCodeImageTooLargeException:
			log.Error().Err(err).Msgf("%v, %v", rekognition.ErrCodeImageTooLargeException, aerr.Error())
		case rekognition.ErrCodeAccessDeniedException:
			log.Error().Err(err).Msgf("%v, %v", rekognition.ErrCodeAccessDeniedException, aerr.Error())
		case rekognition.ErrCodeInternalServerError:
			log.Error().Err(err).Msgf("%v, %v", rekognition.ErrCodeInternalServerError, aerr.Error())
		case rekognition.ErrCodeThrottlingException:
			log.Error().Err(err).Msgf("%v, %v", rekognition.ErrCodeThrottlingException, aerr.Error())
		case rekognition.ErrCodeProvisionedThroughputExceededException:
			log.Error().Err(err).Msgf("%v, %v", rekognition.ErrCodeProvisionedThroughputExceededException, aerr.Error())
		case rekognition.ErrCodeInvalidImageFormatException:
			log.Error().Err(err).Msgf("%v, %v", rekognition.ErrCodeInvalidImageFormatException, aerr.Error())
		default:
			log.Error().Err(err).Msgf(aerr.Error())
		}
	} else {
		log.Error().Err(err).Msgf(err.Error())
	}
}

// GetPublicURLFromKey generates a public url for an object that is being uploaded to s3
// it uses bucket name and region determined with env variables
func GetPublicURLFromKey(key string) string {
	bucketName := os.Getenv(bucketRef)
	region := os.Getenv(regionRef)
	url := fmt.Sprintf(
		"http://%s.s3-%s.amazonaws.com/%s",
		bucketName,
		region,
		key,
	)
	log.Info().Msgf("generated public url: %s", url)
	return url
}

// GetPublicURLFromRecord generates a public url for an object that is being uploaded to s3
// it uses a record that determines the bucket name, aws region and object key
func GetPublicURLFromRecord(record events.S3EventRecord) string {
	imagePublicURL := fmt.Sprintf("%s.s3-%s.amazonaws.com/%s",
		record.S3.Bucket.Name,
		record.AWSRegion,
		record.S3.Object.Key,
	)
	return imagePublicURL
}

// GetImageNameWithExtension generates an image name including its extension
// it uses a key and appends an extension to it
func GetImageNameWithExtension(key string, ext string) string {
	name := fmt.Sprintf(
		"%s.%s",
		key, ext,
	)
	log.Info().Msgf("generated image name with extension: %s", name)
	return name
}
