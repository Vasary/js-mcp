package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	app "github.com/vasary/job-search-mcp/internal/application"
)

type Server struct {
	httpServer *http.Server
	service    *app.Service
	metrics    *metrics
	logger     *slog.Logger
}

func New(addr string, service *app.Service, registry *prometheus.Registry) *Server {
	server := &Server{
		service: service,
		metrics: newMetrics(registry),
		logger:  slog.Default().With("component", "http"),
	}

	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	mux.Handle("GET /healthz", server.instrument("GET", "/healthz", http.HandlerFunc(server.handleHealth)))
	mux.Handle("GET /api/v1/applications", server.instrument("GET", "/api/v1/applications", http.HandlerFunc(server.handleListApplications)))
	mux.Handle("POST /api/v1/applications", server.instrument("POST", "/api/v1/applications", http.HandlerFunc(server.handleCreateApplication)))
	mux.Handle("GET /api/v1/applications/{id}", server.instrument("GET", "/api/v1/applications/{id}", http.HandlerFunc(server.handleGetApplication)))
	mux.Handle("PATCH /api/v1/applications/{id}", server.instrument("PATCH", "/api/v1/applications/{id}", http.HandlerFunc(server.handleUpdateApplication)))
	mux.Handle("POST /api/v1/applications/{id}/comments", server.instrument("POST", "/api/v1/applications/{id}/comments", http.HandlerFunc(server.handleAddComment)))
	mux.Handle("POST /api/v1/applications/{id}/status-changes", server.instrument("POST", "/api/v1/applications/{id}/status-changes", http.HandlerFunc(server.handleChangeStatus)))
	mux.Handle("POST /api/v1/applications/{id}/documents/cv", server.instrument("POST", "/api/v1/applications/{id}/documents/cv", http.HandlerFunc(server.handleUploadCV)))
	mux.Handle("POST /api/v1/applications/{id}/documents/cover-letter", server.instrument("POST", "/api/v1/applications/{id}/documents/cover-letter", http.HandlerFunc(server.handleUploadCoverLetter)))

	server.httpServer = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return server
}

func (s *Server) Run() error {
	err := s.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
