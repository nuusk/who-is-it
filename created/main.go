package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/rs/zerolog/log"
)

type Celebrity struct {
	Name 	string
	ID 		string
}

// CreatedHandler ...
func CreatedHandler(ctx context.Context, event events.S3Event) (bool, error) {
	svc := rekognition.New(session.New())

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

		for _, celeb := range celebRes.CelebrityFaces {
			log.Info().Msgf("%s found", celeb.Name)
		}

		nbNoCelebs := len(celebRes.UnrecognizedFaces)
		log.Info().Msgf("%v unrecognized people", nbNoCelebs)
		// var  Celebrity
		// if err = json.Unmarshal(celebRes, )
		// log.Info().Msgf("%v", celebRes)
	}

	return true, nil
}

func main() {
	lambda.Start(CreatedHandler)
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
