# Job Search MCP Server

Job Search MCP Server is a single Go binary that exposes job application workflows over MCP using the `stdio` transport.

It is intended to be launched directly by an MCP client or agent runtime on the same machine.

## What It Does

The server provides MCP tools for:

- creating tracked job applications
- updating tracked job applications
- listing and searching tracked job applications
- retrieving a single tracked job application with full history
- adding timestamped notes
- changing hiring status with history
- attaching CV PDF files from disk
- attaching cover letter PDF files from disk

## Transport

This project uses:

- MCP over `stdio`

It does not expose REST endpoints and it does not expose MCP over HTTP.

## Data Model

The server stores data in PostgreSQL using a normalized schema with:

- `job_applications`
- `job_application_status_history`
- `job_application_comments`
- `job_application_documents`

Default schema:

- `public`

Detailed schema notes are available in [docs/database.md](docs/database.md).

## Requirements

- Go `1.24`
- PostgreSQL
- a writable directory for uploaded files

## Configuration

### Required

- `JOB_SEARCH_DATABASE_URL`

Example:

```bash
export JOB_SEARCH_DATABASE_URL='postgres://user:password@db-host:5432/database_name'
```

### Optional

- `JOB_SEARCH_FILE_DIR`
  Directory for uploaded files, default: `./var/data`
- `JOB_SEARCH_DB_SCHEMA`
  Database schema, default: `public`
- `JOB_SEARCH_DB_TABLE_APPLICATIONS`
  Default: `job_applications`
- `JOB_SEARCH_DB_TABLE_STATUS_HISTORY`
  Default: `job_application_status_history`
- `JOB_SEARCH_DB_TABLE_COMMENTS`
  Default: `job_application_comments`
- `JOB_SEARCH_DB_TABLE_DOCUMENTS`
  Default: `job_application_documents`

## Running

### With Go

```bash
export JOB_SEARCH_DATABASE_URL='postgres://user:password@db-host:5432/database_name'
go run ./cmd/job-search-mcp
```

### With Make

```bash
export JOB_SEARCH_DATABASE_URL='postgres://user:password@db-host:5432/database_name'
make run
```

### Build Binary

```bash
make build
```

The binary will be created at:

```text
bin/job-search-mcp
```

## MCP Tools

Available tools:

- `create_job_application`
- `update_job_application`
- `list_job_applications`
- `search_job_applications`
- `list_recent_job_applications`
- `get_job_application`
- `get_job_application_timeline`
- `list_job_application_documents`
- `get_job_application_stats`
- `add_job_application_note`
- `change_job_application_status`
- `attach_cv_to_job_application`
- `attach_cover_letter_to_job_application`

## Typical MCP Client Configuration

Example shape:

```json
{
  "mcpServers": {
    "job-search": {
      "command": "/absolute/path/to/job-search-mcp",
      "env": {
        "JOB_SEARCH_DATABASE_URL": "postgres://user:password@db-host:5432/database_name",
        "JOB_SEARCH_FILE_DIR": "/absolute/path/to/data"
      }
    }
  }
}
```

## Development

Run tests:

```bash
make test
```

Build the binary:

```bash
make build
```

## Notes

- The MCP transport is `stdio`, so the server should be launched by the client process
- Logging is intentionally minimal because stdout must stay clean for MCP messages
- Uploaded documents currently support PDF files only
- CV and cover letter files are stored on disk, while PostgreSQL stores metadata
