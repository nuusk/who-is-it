# Who is it
![status](https://github.com/pietersweter/who-is-it/workflows/deploy/badge.svg)

Find out what famous people are on your photos

## Stack:
- [Golang](https://golang.org/)
- [Serverless Framework](https://www.serverless.com/)
- [Amazon S3](https://aws.amazon.com/s3/)
- [Amazon DynamoDB](https://aws.amazon.com/dynamodb/)
- [Amazon SQS](https://aws.amazon.com/sqs/)

## CI/CD
- code is [deployed](https://github.com/pietersweter/who-is-it/actions) to *aws* with each push to the `master` branch (done via [Github Actions](https://github.com/features/actions))

## Endpoints
Upload your images using to `/dev/celeb` endpoint
#### POST `/celeb`
- uploads image to the s3 and runs the celebrity recognition event. Body structure:
  ```
  {
    imageBase64: string,
    fileName: string,
    extension: string
   }
  ```
- this request will return an `url` to an object created in the s3 that will be used later
- related serverless functions:
  - `image-upload` - executed by this request
![](https://pieterweter-repository-images.s3-eu-west-1.amazonaws.com/Screenshot+2020-10-11+at+20.37.25.png) 
  - `image-created` - event triggered by *sqs*

#### GET `/celeb
- returns list of celebrities recognized from your photos. Each celebrity contains an array of photos representing him (that were uploaded to s3)
- related serverless functions:
  - `get-celebs` - executed by this request

## Demo
```
./scripts/batch_send_directory.sh ${API}/celeb ${PHOTOS_LIBRARY}
```
for the sake of testing, you can use `./images` as your `PHOTO_LIBRARY`
