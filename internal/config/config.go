package config

import (
	"errors"
	"os"
)

type Config struct {
	DatabaseURL string
	FileDir     string
	DBSchema    string
	DBTables    DBTables
}

type DBTables struct {
	Applications  string
	StatusHistory string
	Comments      string
	Documents     string
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		DatabaseURL: os.Getenv("JOB_SEARCH_DATABASE_URL"),
		FileDir:     envOrDefault("JOB_SEARCH_FILE_DIR", "./var/data"),
		DBSchema:    envOrDefault("JOB_SEARCH_DB_SCHEMA", "public"),
		DBTables: DBTables{
			Applications:  envOrDefault("JOB_SEARCH_DB_TABLE_APPLICATIONS", "job_applications"),
			StatusHistory: envOrDefault("JOB_SEARCH_DB_TABLE_STATUS_HISTORY", "job_application_status_history"),
			Comments:      envOrDefault("JOB_SEARCH_DB_TABLE_COMMENTS", "job_application_comments"),
			Documents:     envOrDefault("JOB_SEARCH_DB_TABLE_DOCUMENTS", "job_application_documents"),
		},
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("JOB_SEARCH_DATABASE_URL is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
