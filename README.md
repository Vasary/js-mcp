# Job Search Service

Job Search Service is a Go microservice for tracking job applications and exposing that data to both humans and AI agents.

It provides:

- a normalized PostgreSQL-backed data model
- an HTTP API for health checks, metrics, CRUD-like operations, status history, comments, and document uploads
- an MCP server over `stdio` so an AI agent can manage applications directly
- PDF document attachment support for CVs and cover letters
- Prometheus metrics
- structured JSON logging

## Features

- Create and update job applications
- Track immutable status transitions with timestamps
- Store immutable comments with timestamps
- Attach PDF CV files
- Attach PDF cover letter files
- Query a full application including history and uploaded documents
- Expose `healthz` and Prometheus `metrics`
- Run as a small static container image

## Architecture

The service follows a simple layered structure:

- `internal/application`
  Domain models and use cases
- `internal/storage`
  PostgreSQL repository and local file storage for documents
- `internal/httpapi`
  HTTP transport
- `internal/mcpserver`
  MCP transport over `stdio`
- `cmd/job-search-mcp`
  Service bootstrap

Both HTTP and MCP use the same application service layer, so business logic stays in one place.

## Data Model

The PostgreSQL schema is normalized and split into four tables.

Default names are:

- `job_applications`
  Core application fields
- `job_application_status_history`
  Immutable status changes with `changed_at`
- `job_application_comments`
  Immutable comments with `created_at`
- `job_application_documents`
  Metadata for uploaded files such as CVs and cover letters

Schema and table names are configurable at runtime.

Detailed database documentation is available in [docs/database.md](docs/database.md).

SQL migrations are stored in:

- [migrations/0001_job_search_service.up.sql](migrations/0001_job_search_service.up.sql)
- [migrations/0001_job_search_service.down.sql](migrations/0001_job_search_service.down.sql)

## Application Statuses

Supported statuses:

- `applied`
- `screening`
- `interview`
- `offer`
- `rejected`
- `withdrawn`
- `accepted`

## Supported Document Types

- `cv`
- `cover_letter`

Both document types are validated as PDF files before they are stored.

## Requirements

- Go `1.24`
- PostgreSQL
- A writable directory for uploaded files

## Configuration

The service is configured through environment variables.

### Required

- `JOB_SEARCH_DATABASE_URL`

Example:

```bash
export JOB_SEARCH_DATABASE_URL='postgres://user:password@db-host:5432/database_name'
```

### Optional

- `JOB_SEARCH_HTTP_ADDR`
  HTTP listen address, default: `:8080`
- `JOB_SEARCH_FILE_DIR`
  Root directory for uploaded files, default: `./var/data`
- `JOB_SEARCH_ENABLE_MCP`
  Enable MCP server on `stdio`, default: `false`
- `JOB_SEARCH_LOG_LEVEL`
  `info` or `debug`, default: `info`
- `JOB_SEARCH_DB_SCHEMA`
  Database schema name, default: `openclaw`
- `JOB_SEARCH_DB_TABLE_APPLICATIONS`
  Applications table name, default: `job_applications`
- `JOB_SEARCH_DB_TABLE_STATUS_HISTORY`
  Status history table name, default: `job_application_status_history`
- `JOB_SEARCH_DB_TABLE_COMMENTS`
  Comments table name, default: `job_application_comments`
- `JOB_SEARCH_DB_TABLE_DOCUMENTS`
  Documents table name, default: `job_application_documents`

## Running Locally

### With Make

```bash
export JOB_SEARCH_DATABASE_URL='postgres://user:password@db-host:5432/database_name'
make run
```

### Directly with Go

```bash
export JOB_SEARCH_DATABASE_URL='postgres://user:password@db-host:5432/database_name'
go run ./cmd/job-search-mcp
```

### Run Tests

```bash
make test
```

### Build Binary

```bash
make build
```

## Docker

The project includes a multi-stage Docker build that produces a very small `scratch` runtime image.

### Build Image

```bash
make docker-build
```

Or directly:

```bash
docker build -t job-search-service:local .
```

### Notes About the Container

- The binary is built statically with `CGO_ENABLED=0`
- The runtime image is based on `scratch`
- You must provide:
  - database connection settings
  - a writable volume for uploaded files

Example:

```bash
docker run --rm \
  -p 8080:8080 \
  -e JOB_SEARCH_DATABASE_URL='postgres://user:password@db-host:5432/database_name' \
  -e JOB_SEARCH_FILE_DIR='/data' \
  -v $(pwd)/var/data:/data \
  job-search-service:local
```

## HTTP API

### Health and Metrics

- `GET /healthz`
- `GET /metrics`

### Applications

- `GET /api/v1/applications`
- `POST /api/v1/applications`
- `GET /api/v1/applications/{id}`
- `PATCH /api/v1/applications/{id}`

### Comments

- `POST /api/v1/applications/{id}/comments`

### Status History

- `POST /api/v1/applications/{id}/status-changes`

### Documents

- `POST /api/v1/applications/{id}/documents/cv`
- `POST /api/v1/applications/{id}/documents/cover-letter`

### Example: Create Application

```bash
curl -X POST http://localhost:8080/api/v1/applications \
  -H 'Content-Type: application/json' \
  -d '{
    "companyName": "OpenAI",
    "positionTitle": "Backend Engineer",
    "sourceUrl": "https://example.com/jobs/backend-engineer",
    "workType": "remote",
    "salary": "$150k-$180k",
    "positionDescription": "Backend role focused on APIs and distributed systems",
    "techStack": "Go, PostgreSQL, Docker",
    "initialStatus": "applied",
    "initialStatusNote": "Applied through careers page",
    "initialComment": "Strong fit for backend platform work"
  }'
```

### Example: Add Comment

```bash
curl -X POST http://localhost:8080/api/v1/applications/1/comments \
  -H 'Content-Type: application/json' \
  -d '{
    "body": "Recruiter responded and asked for availability."
  }'
```

### Example: Change Status

```bash
curl -X POST http://localhost:8080/api/v1/applications/1/status-changes \
  -H 'Content-Type: application/json' \
  -d '{
    "status": "interview",
    "note": "Technical interview scheduled for next week"
  }'
```

### Example: Upload CV

```bash
curl -X POST http://localhost:8080/api/v1/applications/1/documents/cv \
  -F 'file=@./resume.pdf'
```

### Example: Upload Cover Letter

```bash
curl -X POST http://localhost:8080/api/v1/applications/1/documents/cover-letter \
  -F 'file=@./cover-letter.pdf'
```

## MCP Interface

When `JOB_SEARCH_ENABLE_MCP=true`, the service also exposes an MCP server over standard input/output.

Available MCP tools:

- `create_application`
- `update_application`
- `list_applications`
- `get_application`
- `add_comment`
- `change_status`
- `upload_cv_from_path`
- `upload_cover_letter_from_path`

### MCP Use Cases

This lets an AI assistant:

- create new applications
- inspect application history
- add comments after conversations
- update current status
- attach local CV and cover letter PDF files

## Logging

The service uses structured JSON logging via Go `slog`.

Logged events include:

- startup and shutdown
- PostgreSQL readiness
- HTTP request summaries
- fatal bootstrap failures

Set:

```bash
export JOB_SEARCH_LOG_LEVEL=debug
```

for more verbose logs.

## Prometheus Metrics

The service exposes Prometheus metrics on:

```text
/metrics
```

Currently included:

- HTTP request count
- HTTP request duration
- Go runtime metrics
- process metrics

## File Storage

Uploaded documents are stored on local disk under the configured root directory.

Example layout:

```text
var/data/
  applications/
    42/
      cv/
        7e0a...d9.pdf
      cover_letter/
        5ac1...3f.pdf
```

The database stores only metadata:

- original filename
- content type
- storage path
- file hash
- size
- upload timestamp

## Migration Notes

The service code expects the normalized schema from `migrations/0001_job_search_service.up.sql`.

If you are migrating from the previous single-table version, apply the new migration before starting the service.

## Development Notes

- Business logic is transport-agnostic
- HTTP and MCP are thin wrappers over the same use cases
- Status history and comments are append-only by design
- Document uploads currently support PDF only

## Future Improvements

Possible next steps:

- migration runner integration
- document download endpoints
- authentication and authorization
- object storage backend for documents
- OpenAPI specification
- pagination metadata improvements
