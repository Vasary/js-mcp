APP_NAME=job-search-service
IMAGE_NAME?=job-search-service:local
HTTP_ADDR?=:8080
FILE_DIR?=./var/data

.PHONY: run test build docker-build

run:
	JOB_SEARCH_HTTP_ADDR=$(HTTP_ADDR) \
	JOB_SEARCH_FILE_DIR=$(FILE_DIR) \
	go run ./cmd/job-search-mcp

test:
	GOCACHE=$(CURDIR)/.gocache GOPROXY=off GOSUMDB=off go test ./...

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/$(APP_NAME) ./cmd/job-search-mcp

docker-build:
	docker build -t $(IMAGE_NAME) .
