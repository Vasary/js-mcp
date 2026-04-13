package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	app "github.com/vasary/job-search-mcp/internal/application"
)

func translateNotFound(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return app.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return app.ErrNotFound
	}
	return err
}

func nullIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (r *PostgresRepository) ensureApplicationExists(ctx context.Context, applicationID int64) error {
	query := fmt.Sprintf(`select 1 from %s where id = $1`, r.applicationsTable())
	var one int
	err := r.pool.QueryRow(ctx, query, applicationID).Scan(&one)
	return translateNotFound(err)
}
