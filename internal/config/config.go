package config

import (
	"errors"
	"os"
)

type Config struct {
	DatabaseURL    string
	HTTPAddr       string
	FileDir        string
	EnableMCP      bool
	LogLevel       string
	HealthcheckURL string
	DBSchema       string
	DBTables       DBTables
}

type DBTables struct {
	Applications  string
	StatusHistory string
	Comments      string
	Documents     string
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		DatabaseURL:    os.Getenv("JOB_SEARCH_DATABASE_URL"),
		HTTPAddr:       envOrDefault("JOB_SEARCH_HTTP_ADDR", ":80"),
		FileDir:        envOrDefault("JOB_SEARCH_FILE_DIR", "/tmp"),
		EnableMCP:      os.Getenv("JOB_SEARCH_ENABLE_MCP") == "true",
		LogLevel:       envOrDefault("JOB_SEARCH_LOG_LEVEL", "info"),
		HealthcheckURL: envOrDefault("JOB_SEARCH_HEALTHCHECK_URL", "http://127.0.0.1:8080/healthz"),
		DBSchema:       envOrDefault("JOB_SEARCH_DB_SCHEMA", "openclaw"),
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
