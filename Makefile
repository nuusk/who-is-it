.PHONY: build clean deploy

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/image-upload handlers/image-upload/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/image-created handlers/image-created/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/get-celebs handlers/get-celebs/main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
