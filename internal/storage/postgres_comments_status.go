package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	app "github.com/vasary/job-search-mcp/internal/application"
)

func (r *PostgresRepository) AddComment(ctx context.Context, input app.AddCommentInput, now time.Time) (app.Comment, error) {
	comment, err := r.insertComment(ctx, r.pool, input, now)
	if err != nil {
		return app.Comment{}, err
	}
	if err := r.touchApplication(ctx, input.ApplicationID, now); err != nil {
		return app.Comment{}, err
	}
	return comment, nil
}

func (r *PostgresRepository) ChangeStatus(ctx context.Context, input app.ChangeStatusInput, now time.Time) (app.StatusChange, error) {
	change, err := r.insertStatusChange(ctx, r.pool, input, now)
	if err != nil {
		return app.StatusChange{}, err
	}
	if err := r.touchApplication(ctx, input.ApplicationID, now); err != nil {
		return app.StatusChange{}, err
	}
	return change, nil
}

func (r *PostgresRepository) insertComment(ctx context.Context, q queryRower, input app.AddCommentInput, now time.Time) (app.Comment, error) {
	const query = `
		insert into %s (
			application_id,
			body,
			created_at
		)
		values ($1, $2, $3)
		returning id, application_id, body, created_at
	`

	var comment app.Comment
	err := q.QueryRow(ctx, fmt.Sprintf(query, r.commentsTable()), input.ApplicationID, input.Body, now).Scan(
		&comment.ID,
		&comment.ApplicationID,
		&comment.Body,
		&comment.CreatedAt,
	)
	return comment, translateNotFound(err)
}

func (r *PostgresRepository) insertStatusChange(ctx context.Context, q queryRower, input app.ChangeStatusInput, now time.Time) (app.StatusChange, error) {
	const query = `
		insert into %s (
			application_id,
			status,
			note,
			changed_at
		)
		values ($1, $2, $3, $4)
		returning id, application_id, status, note, changed_at
	`

	var (
		change app.StatusChange
		status string
		note   *string
	)
	err := q.QueryRow(ctx, fmt.Sprintf(query, r.statusHistoryTable()), input.ApplicationID, input.Status, nullIfEmpty(input.Note), now).Scan(
		&change.ID,
		&change.ApplicationID,
		&status,
		&note,
		&change.ChangedAt,
	)
	change.Status = app.ApplicationStatus(status)
	change.Note = derefString(note)
	return change, translateNotFound(err)
}

func (r *PostgresRepository) listComments(ctx context.Context, applicationID int64) ([]app.Comment, error) {
	const query = `
		select id, application_id, body, created_at
		from %s
		where application_id = $1
		order by created_at asc, id asc
	`

	rows, err := r.pool.Query(ctx, fmt.Sprintf(query, r.commentsTable()), applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []app.Comment
	for rows.Next() {
		var item app.Comment
		if err := rows.Scan(&item.ID, &item.ApplicationID, &item.Body, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) listStatusHistory(ctx context.Context, applicationID int64) ([]app.StatusChange, error) {
	const query = `
		select id, application_id, status, note, changed_at
		from %s
		where application_id = $1
		order by changed_at desc, id desc
	`

	rows, err := r.pool.Query(ctx, fmt.Sprintf(query, r.statusHistoryTable()), applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []app.StatusChange
	for rows.Next() {
		var (
			item   app.StatusChange
			status string
			note   *string
		)
		if err := rows.Scan(&item.ID, &item.ApplicationID, &status, &note, &item.ChangedAt); err != nil {
			return nil, err
		}
		item.Status = app.ApplicationStatus(status)
		item.Note = derefString(note)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *PostgresRepository) touchApplication(ctx context.Context, applicationID int64, now time.Time) error {
	query := fmt.Sprintf(`update %s set updated_at = $2 where id = $1`, r.applicationsTable())
	tag, err := r.pool.Exec(ctx, query, applicationID, now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return app.ErrNotFound
	}
	return nil
}

type queryRower interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}
