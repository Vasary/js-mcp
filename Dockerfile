FROM golang:1.24 AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/job-search-service ./cmd/job-search-mcp

FROM scratch

WORKDIR /app

COPY --from=build /out/job-search-service /job-search-service

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD ["/job-search-service", "healthcheck"]

ENTRYPOINT ["/job-search-service"]
