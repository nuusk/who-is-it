.PHONY: build clean deploy

build:
	env GOOS=linux go build -ldflags="-s -w" -o bin/upload upload/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/list list/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/created created/main.go

clean:
	rm -rf ./bin

deploy: clean build
	sls deploy --verbose
