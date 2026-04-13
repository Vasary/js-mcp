package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	app "github.com/vasary/job-search-mcp/internal/application"
	"github.com/vasary/job-search-mcp/internal/config"
	"github.com/vasary/job-search-mcp/internal/httpapi"
	"github.com/vasary/job-search-mcp/internal/mcpserver"
	"github.com/vasary/job-search-mcp/internal/observability"
	"github.com/vasary/job-search-mcp/internal/storage"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		if err := runHealthcheck(); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.LoadFromEnv()
	if err != nil {
		return err
	}

	logger := observability.NewLogger(cfg.LogLevel)
	slog.SetDefault(logger)
	logger.Info("starting service", "http_addr", cfg.HTTPAddr, "file_root", cfg.FileDir, "mcp_enabled", cfg.EnableMCP)

	service, cleanup, err := buildService(ctx, cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	if err := service.Health(ctx); err != nil {
		logger.Error("startup health check failed", "error", err)
		return err
	}
	logger.Info("postgres connection established")

	httpServer := httpapi.New(cfg.HTTPAddr, service, observability.NewRegistry())
	errCh := make(chan error, 2)

	go func() {
		logger.Info("http server listening", "addr", cfg.HTTPAddr)
		errCh <- httpServer.Run()
	}()

	if cfg.EnableMCP {
		go func() {
			logger.Info("mcp server enabled on stdio")
			errCh <- mcpserver.New(service).Run(ctx)
		}()
	}

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("server terminated with error", "error", err)
			return err
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http shutdown failed", "error", err)
		return err
	}
	logger.Info("service stopped")
	return nil
}

func runHealthcheck() error {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		// For a container healthcheck we only need the target URL, not the DB settings.
		cfg.HealthcheckURL = configHealthcheckURL()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.HealthcheckURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("healthcheck failed")
	}

	return nil
}

func buildService(ctx context.Context, cfg config.Config) (*app.Service, func(), error) {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Default().Error("failed to create postgres pool", "error", err)
		return nil, nil, err
	}

	repo := storage.NewPostgresRepositoryWithTables(pool, storage.TableNames{
		Schema:        cfg.DBSchema,
		Applications:  cfg.DBTables.Applications,
		StatusHistory: cfg.DBTables.StatusHistory,
		Comments:      cfg.DBTables.Comments,
		Documents:     cfg.DBTables.Documents,
	})
	files := storage.NewLocalFileStore(cfg.FileDir, app.DefaultUploadSize)
	service := app.NewService(repo, files)

	cleanup := func() {
		pool.Close()
	}

	return service, cleanup, nil
}

func configHealthcheckURL() string {
	if value := os.Getenv("JOB_SEARCH_HEALTHCHECK_URL"); value != "" {
		return value
	}
	return "http://127.0.0.1:8080/healthz"
}
