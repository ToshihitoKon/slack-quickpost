help:
	cat Makefile | grep '^\w'

build:
	go build -o bin/slack-quickpost

run:
	go run ./...
