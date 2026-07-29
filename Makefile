help:
	cat Makefile | grep '^\w'

build:
	go build -ldflags="-s -w" -o bin/slack-quickpost .

run:
	go run .
