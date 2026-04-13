APP_NAME=job-search-mcp
FILE_DIR?=./var/data

.PHONY: run test build

run:
	JOB_SEARCH_FILE_DIR=$(FILE_DIR) \
	go run ./cmd/job-search-mcp

test:
	GOCACHE=$(CURDIR)/.gocache GOPROXY=off GOSUMDB=off go test ./...

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/$(APP_NAME) ./cmd/job-search-mcp
