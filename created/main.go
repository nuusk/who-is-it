package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/rs/zerolog/log"
)

// CreatedHandler ...
func CreatedHandler(ctx context.Context, event events.S3Event) (bool, error) {
	svc := rekognition.New(session.New())

	for _, record := range event.Records {
		log.Info().Msgf("%v\n %v", record.S3.Bucket.Name, record.S3.Object.Key)

		input := &rekognition.DetectLabelsInput{
			Image: &rekognition.Image{
				S3Object: &rekognition.S3Object{
					Bucket: aws.String(record.S3.Bucket.Name),
					Name:   aws.String(record.S3.Object.Key),
				},
			},
			MaxLabels:     aws.Int64(12),
			MinConfidence: aws.Float64(70.000000),
		}

		result, err := svc.DetectLabels(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case rekognition.ErrCodeInvalidS3ObjectException:
					fmt.Println(rekognition.ErrCodeInvalidS3ObjectException, aerr.Error())
				case rekognition.ErrCodeInvalidParameterException:
					fmt.Println(rekognition.ErrCodeInvalidParameterException, aerr.Error())
				case rekognition.ErrCodeImageTooLargeException:
					fmt.Println(rekognition.ErrCodeImageTooLargeException, aerr.Error())
				case rekognition.ErrCodeAccessDeniedException:
					fmt.Println(rekognition.ErrCodeAccessDeniedException, aerr.Error())
				case rekognition.ErrCodeInternalServerError:
					fmt.Println(rekognition.ErrCodeInternalServerError, aerr.Error())
				case rekognition.ErrCodeThrottlingException:
					fmt.Println(rekognition.ErrCodeThrottlingException, aerr.Error())
				case rekognition.ErrCodeProvisionedThroughputExceededException:
					fmt.Println(rekognition.ErrCodeProvisionedThroughputExceededException, aerr.Error())
				case rekognition.ErrCodeInvalidImageFormatException:
					fmt.Println(rekognition.ErrCodeInvalidImageFormatException, aerr.Error())
				default:
					fmt.Println(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				fmt.Println(err.Error())
			}
			return false, nil
		}

		log.Info().Msgf("%v", result)
	}

	return true, nil
}

func main() {
	lambda.Start(CreatedHandler)
}
