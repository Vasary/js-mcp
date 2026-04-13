package httpapi

import (
	"context"
	"io"
	"net/http"

	app "github.com/vasary/job-search-mcp/internal/application"
)

func uploadDocument(w http.ResponseWriter, r *http.Request, uploader func(context.Context, int64, string, io.Reader) (app.Document, error)) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := r.ParseMultipartForm(app.DefaultUploadSize + (1 << 20)); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	defer file.Close()

	result, err := uploader(r.Context(), id, header.Filename, file)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}
