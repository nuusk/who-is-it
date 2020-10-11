// Package awshelpers provides helpers and handlers for communication and interacting with s3, dynamoDB, rekognition
package awshelpers

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/rs/zerolog/log"
)

func handleDynamoDBError(err error) {
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
