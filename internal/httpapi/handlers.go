package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	app "github.com/vasary/job-search-mcp/internal/application"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := s.service.Health(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (s *Server) handleCreateApplication(w http.ResponseWriter, r *http.Request) {
	var input app.CreateApplicationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result, err := s.service.CreateApplication(r.Context(), input)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) handleListApplications(w http.ResponseWriter, r *http.Request) {
	input := app.ListApplicationsInput{
		CompanyName:   r.URL.Query().Get("companyName"),
		PositionTitle: r.URL.Query().Get("positionTitle"),
		CurrentStatus: app.ApplicationStatus(r.URL.Query().Get("currentStatus")),
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		value, err := strconv.Atoi(limit)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		input.Limit = value
	}
	if offset := r.URL.Query().Get("offset"); offset != "" {
		value, err := strconv.Atoi(offset)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		input.Offset = value
	}

	result, err := s.service.ListApplications(r.Context(), input)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetApplication(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result, err := s.service.GetApplication(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleUpdateApplication(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var input app.UpdateApplicationInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	input.ID = id

	result, err := s.service.UpdateApplication(r.Context(), input)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleAddComment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var input struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result, err := s.service.AddComment(r.Context(), app.AddCommentInput{
		ApplicationID: id,
		Body:          input.Body,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) handleChangeStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var input struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	result, err := s.service.ChangeStatus(r.Context(), app.ChangeStatusInput{
		ApplicationID: id,
		Status:        app.ApplicationStatus(input.Status),
		Note:          input.Note,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) handleUploadCV(w http.ResponseWriter, r *http.Request) {
	uploadDocument(w, r, func(ctx context.Context, applicationID int64, filename string, body io.Reader) (app.Document, error) {
		return s.service.UploadCV(ctx, app.UploadCVInput{
			ApplicationID: applicationID,
			Filename:      filename,
			Body:          body,
		})
	})
}

func (s *Server) handleUploadCoverLetter(w http.ResponseWriter, r *http.Request) {
	uploadDocument(w, r, func(ctx context.Context, applicationID int64, filename string, body io.Reader) (app.Document, error) {
		return s.service.UploadCoverLetter(ctx, app.UploadCoverLetterInput{
			ApplicationID: applicationID,
			Filename:      filename,
			Body:          body,
		})
	})
}
