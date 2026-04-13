package storage

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool   *pgxpool.Pool
	tables TableNames
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return NewPostgresRepositoryWithTables(pool, DefaultTableNames())
}

func NewPostgresRepositoryWithTables(pool *pgxpool.Pool, tables TableNames) *PostgresRepository {
	return &PostgresRepository{
		pool:   pool,
		tables: tables.withDefaults(),
	}
}

type TableNames struct {
	Schema        string
	Applications  string
	StatusHistory string
	Comments      string
	Documents     string
}

func DefaultTableNames() TableNames {
	return TableNames{
		Schema:        "openclaw",
		Applications:  "job_applications",
		StatusHistory: "job_application_status_history",
		Comments:      "job_application_comments",
		Documents:     "job_application_documents",
	}
}

func (t TableNames) withDefaults() TableNames {
	def := DefaultTableNames()
	if t.Schema == "" {
		t.Schema = def.Schema
	}
	if t.Applications == "" {
		t.Applications = def.Applications
	}
	if t.StatusHistory == "" {
		t.StatusHistory = def.StatusHistory
	}
	if t.Comments == "" {
		t.Comments = def.Comments
	}
	if t.Documents == "" {
		t.Documents = def.Documents
	}
	return t
}

func (r *PostgresRepository) applicationsTable() string {
	return pgx.Identifier{r.tables.Schema, r.tables.Applications}.Sanitize()
}

func (r *PostgresRepository) statusHistoryTable() string {
	return pgx.Identifier{r.tables.Schema, r.tables.StatusHistory}.Sanitize()
}

func (r *PostgresRepository) commentsTable() string {
	return pgx.Identifier{r.tables.Schema, r.tables.Comments}.Sanitize()
}

func (r *PostgresRepository) documentsTable() string {
	return pgx.Identifier{r.tables.Schema, r.tables.Documents}.Sanitize()
}
