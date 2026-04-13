package storage

import (
	"context"
	"fmt"
	"time"

	app "github.com/vasary/job-search-mcp/internal/application"
)

func (r *PostgresRepository) AddDocument(ctx context.Context, input app.AddDocumentRecordInput, now time.Time) (app.Document, error) {
	const query = `
		insert into %s (
			application_id,
			document_type,
			original_filename,
			content_type,
			storage_path,
			sha256,
			size_bytes,
			uploaded_at
		)
		values ($1, $2, $3, $4, $5, $6, $7, $8)
		returning
			id,
			application_id,
			document_type,
			original_filename,
			content_type,
			storage_path,
			sha256,
			size_bytes,
			uploaded_at
	`

	var document app.Document
	err := r.pool.QueryRow(
		ctx,
		fmt.Sprintf(query, r.documentsTable()),
		input.ApplicationID,
		input.DocumentType,
		input.OriginalFilename,
		input.ContentType,
		input.StoragePath,
		input.SHA256,
		input.SizeBytes,
		now,
	).Scan(
		&document.ID,
		&document.ApplicationID,
		&document.DocumentType,
		&document.OriginalFilename,
		&document.ContentType,
		&document.StoragePath,
		&document.SHA256,
		&document.SizeBytes,
		&document.UploadedAt,
	)
	if err != nil {
		return app.Document{}, err
	}
	if err := r.touchApplication(ctx, input.ApplicationID, now); err != nil {
		return app.Document{}, err
	}
	return document, nil
}

func (r *PostgresRepository) listDocuments(ctx context.Context, applicationID int64) ([]app.Document, error) {
	const query = `
		select
			id,
			application_id,
			document_type,
			original_filename,
			content_type,
			storage_path,
			sha256,
			size_bytes,
			uploaded_at
		from %s
		where application_id = $1
		order by uploaded_at desc, id desc
	`

	rows, err := r.pool.Query(ctx, fmt.Sprintf(query, r.documentsTable()), applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []app.Document
	for rows.Next() {
		var item app.Document
		if err := rows.Scan(
			&item.ID,
			&item.ApplicationID,
			&item.DocumentType,
			&item.OriginalFilename,
			&item.ContentType,
			&item.StoragePath,
			&item.SHA256,
			&item.SizeBytes,
			&item.UploadedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
