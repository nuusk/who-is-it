.PHONY: build clean deploy

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/upload handlers/upload/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/list handlers/list/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/created handlers/created/main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
