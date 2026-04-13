package storage

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	app "github.com/vasary/job-search-mcp/internal/application"
)

type LocalFileStore struct {
	rootDir       string
	maxUploadSize int64
}

func NewLocalFileStore(rootDir string, maxUploadSize int64) *LocalFileStore {
	if maxUploadSize <= 0 {
		maxUploadSize = app.DefaultUploadSize
	}
	return &LocalFileStore{
		rootDir:       rootDir,
		maxUploadSize: maxUploadSize,
	}
}

func (s *LocalFileStore) SaveDocument(_ context.Context, applicationID int64, documentType, originalFilename string, body io.Reader) (app.StoredFile, error) {
	reader := bufio.NewReader(body)
	header, err := reader.Peek(5)
	if err != nil {
		return app.StoredFile{}, err
	}
	if string(header) != "%PDF-" {
		return app.StoredFile{}, app.ErrInvalidPDF
	}

	filename := strings.TrimSpace(originalFilename)
	if filename == "" {
		filename = "cv.pdf"
	}
	if !strings.HasSuffix(strings.ToLower(filename), ".pdf") {
		filename += ".pdf"
	}
	if documentType != app.DocumentTypeCV && documentType != app.DocumentTypeCoverLetter {
		return app.StoredFile{}, app.ErrUnsupportedDoc
	}

	dir := filepath.Join(s.rootDir, "applications", fmt.Sprintf("%d", applicationID), documentType)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return app.StoredFile{}, err
	}

	targetName := randomHex(16) + ".pdf"
	targetPath := filepath.Join(dir, targetName)
	file, err := os.Create(targetPath)
	if err != nil {
		return app.StoredFile{}, err
	}
	defer file.Close()

	hasher := sha256.New()
	written, err := io.Copy(io.MultiWriter(file, hasher), io.LimitReader(reader, s.maxUploadSize+1))
	if err != nil {
		_ = os.Remove(targetPath)
		return app.StoredFile{}, err
	}
	if written > s.maxUploadSize {
		_ = os.Remove(targetPath)
		return app.StoredFile{}, fmt.Errorf("%w: file exceeds %d bytes", app.ErrInvalidInput, s.maxUploadSize)
	}

	return app.StoredFile{
		OriginalFilename: filepath.Base(filename),
		ContentType:      "application/pdf",
		StoragePath:      targetPath,
		SHA256:           hex.EncodeToString(hasher.Sum(nil)),
		SizeBytes:        written,
	}, nil
}

func (s *LocalFileStore) Delete(_ context.Context, storagePath string) error {
	if strings.TrimSpace(storagePath) == "" {
		return nil
	}
	err := os.Remove(storagePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func randomHex(size int) string {
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		panic(err)
	}
	return hex.EncodeToString(raw)
}
