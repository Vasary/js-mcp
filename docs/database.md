# Job Search Service Database

## Overview

The service uses a normalized PostgreSQL schema in the `openclaw` schema.

Main entities:

- `job_applications`: the core application record
- `job_application_status_history`: immutable status transitions with timestamps
- `job_application_comments`: immutable comments with timestamps
- `job_application_documents`: uploaded files such as CV PDFs

This keeps the current application snapshot compact while preserving full history for audit and AI workflows.

## Tables

### `openclaw.job_applications`

Stores the stable fields of an application:

- `company_name`
- `position_title`
- `source_url`
- `work_type`
- `salary`
- `position_description`
- `tech_stack`
- `created_at`
- `updated_at`

### `openclaw.job_application_status_history`

Stores every status change as a separate immutable event:

- `application_id`
- `status`
- `note`
- `changed_at`

The current status is derived from the latest row for an application.

### `openclaw.job_application_comments`

Stores comments independently from the application row:

- `application_id`
- `body`
- `created_at`

This makes comment history explicit and queryable.

### `openclaw.job_application_documents`

Stores metadata for uploaded documents:

- `application_id`
- `document_type`
- `original_filename`
- `content_type`
- `storage_path`
- `sha256`
- `size_bytes`
- `uploaded_at`

Supported document types:

- `cv`
- `cover_letter`

The binary file itself is stored on disk, while PostgreSQL stores only metadata.

## Current Status Query

Current status is read from the latest status history event:

```sql
select status, changed_at
from openclaw.job_application_status_history
where application_id = $1
order by changed_at desc, id desc
limit 1;
```

## Notes

- Statuses are constrained to: `applied`, `screening`, `interview`, `offer`, `rejected`, `withdrawn`, `accepted`
- Documents currently support `cv` and `cover_letter`
- CV and cover letter uploads are validated as PDF files before metadata is written to the database
