package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	app "github.com/vasary/job-search-mcp/internal/application"
	"github.com/vasary/job-search-mcp/internal/config"
	"github.com/vasary/job-search-mcp/internal/mcpserver"
	"github.com/vasary/job-search-mcp/internal/storage"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	repo := storage.NewPostgresRepositoryWithTables(pool, storage.TableNames{
		Schema:        cfg.DBSchema,
		Applications:  cfg.DBTables.Applications,
		StatusHistory: cfg.DBTables.StatusHistory,
		Comments:      cfg.DBTables.Comments,
		Documents:     cfg.DBTables.Documents,
	})
	files := storage.NewLocalFileStore(cfg.FileDir, app.DefaultUploadSize)
	service := app.NewService(repo, files)

	if err := service.Health(ctx); err != nil {
		return err
	}

	return mcpserver.New(service).Run(ctx)
}
