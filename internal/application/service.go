package application

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrNotFound       = errors.New("application not found")
	ErrInvalidStatus  = errors.New("invalid application status")
	ErrInvalidPDF     = errors.New("file must be a PDF")
	ErrInvalidInput   = errors.New("invalid input")
	ErrUnsupportedDoc = errors.New("unsupported document type")
)

type Repository interface {
	Ping(ctx context.Context) error
	CreateApplication(ctx context.Context, input CreateApplicationInput, now time.Time) (ApplicationDetails, error)
	UpdateApplication(ctx context.Context, input UpdateApplicationInput, now time.Time) (ApplicationDetails, error)
	GetApplication(ctx context.Context, id int64) (ApplicationDetails, error)
	ListApplications(ctx context.Context, input ListApplicationsInput) (ListApplicationsOutput, error)
	AddComment(ctx context.Context, input AddCommentInput, now time.Time) (Comment, error)
	ChangeStatus(ctx context.Context, input ChangeStatusInput, now time.Time) (StatusChange, error)
	AddDocument(ctx context.Context, input AddDocumentRecordInput, now time.Time) (Document, error)
}

type FileStore interface {
	SaveDocument(ctx context.Context, applicationID int64, documentType, originalFilename string, body io.Reader) (StoredFile, error)
	Delete(ctx context.Context, storagePath string) error
}

type StoredFile struct {
	OriginalFilename string
	ContentType      string
	StoragePath      string
	SHA256           string
	SizeBytes        int64
}

type Service struct {
	repo  Repository
	files FileStore
	now   func() time.Time
}

func NewService(repo Repository, files FileStore) *Service {
	return &Service{
		repo:  repo,
		files: files,
		now:   func() time.Time { return time.Now().UTC() },
	}
}

type CreateApplicationInput struct {
	CompanyName         string            `json:"companyName"`
	PositionTitle       string            `json:"positionTitle,omitempty"`
	SourceURL           string            `json:"sourceUrl,omitempty"`
	WorkType            string            `json:"workType,omitempty"`
	Salary              string            `json:"salary,omitempty"`
	PositionDescription string            `json:"positionDescription,omitempty"`
	TechStack           string            `json:"techStack,omitempty"`
	InitialStatus       ApplicationStatus `json:"initialStatus,omitempty"`
	InitialStatusNote   string            `json:"initialStatusNote,omitempty"`
	InitialComment      string            `json:"initialComment,omitempty"`
}

type UpdateApplicationInput struct {
	ID                  int64   `json:"id"`
	CompanyName         *string `json:"companyName,omitempty"`
	PositionTitle       *string `json:"positionTitle,omitempty"`
	SourceURL           *string `json:"sourceUrl,omitempty"`
	WorkType            *string `json:"workType,omitempty"`
	Salary              *string `json:"salary,omitempty"`
	PositionDescription *string `json:"positionDescription,omitempty"`
	TechStack           *string `json:"techStack,omitempty"`
}

type ListApplicationsInput struct {
	CompanyName   string            `json:"companyName,omitempty"`
	PositionTitle string            `json:"positionTitle,omitempty"`
	CurrentStatus ApplicationStatus `json:"currentStatus,omitempty"`
	Limit         int               `json:"limit,omitempty"`
	Offset        int               `json:"offset,omitempty"`
}

type ListApplicationsOutput struct {
	Items []ApplicationSummary `json:"items"`
	Total int                  `json:"total"`
}

type AddCommentInput struct {
	ApplicationID int64  `json:"applicationId"`
	Body          string `json:"body"`
}

type ChangeStatusInput struct {
	ApplicationID int64             `json:"applicationId"`
	Status        ApplicationStatus `json:"status"`
	Note          string            `json:"note,omitempty"`
}

type AddDocumentRecordInput struct {
	ApplicationID    int64  `json:"applicationId"`
	DocumentType     string `json:"documentType"`
	OriginalFilename string `json:"originalFilename"`
	ContentType      string `json:"contentType"`
	StoragePath      string `json:"storagePath"`
	SHA256           string `json:"sha256"`
	SizeBytes        int64  `json:"sizeBytes"`
}

type UploadCVInput struct {
	ApplicationID int64
	Filename      string
	Body          io.Reader
}

type UploadCVFromPathInput struct {
	ApplicationID int64  `json:"applicationId"`
	FilePath      string `json:"filePath"`
}

type UploadCoverLetterInput struct {
	ApplicationID int64
	Filename      string
	Body          io.Reader
}

type UploadCoverLetterFromPathInput struct {
	ApplicationID int64  `json:"applicationId"`
	FilePath      string `json:"filePath"`
}

func (s *Service) Health(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

func (s *Service) CreateApplication(ctx context.Context, input CreateApplicationInput) (ApplicationDetails, error) {
	input.CompanyName = strings.TrimSpace(input.CompanyName)
	input.PositionTitle = strings.TrimSpace(input.PositionTitle)
	input.SourceURL = strings.TrimSpace(input.SourceURL)
	input.WorkType = strings.TrimSpace(input.WorkType)
	input.Salary = strings.TrimSpace(input.Salary)
	input.PositionDescription = strings.TrimSpace(input.PositionDescription)
	input.TechStack = strings.TrimSpace(input.TechStack)
	input.InitialStatusNote = strings.TrimSpace(input.InitialStatusNote)
	input.InitialComment = strings.TrimSpace(input.InitialComment)

	if input.CompanyName == "" {
		return ApplicationDetails{}, fmt.Errorf("%w: companyName is required", ErrInvalidInput)
	}

	if input.InitialStatus == "" {
		input.InitialStatus = StatusApplied
	}
	if !input.InitialStatus.Valid() {
		return ApplicationDetails{}, ErrInvalidStatus
	}

	return s.repo.CreateApplication(ctx, input, s.now())
}

func (s *Service) UpdateApplication(ctx context.Context, input UpdateApplicationInput) (ApplicationDetails, error) {
	if input.ID <= 0 {
		return ApplicationDetails{}, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}
	if input.CompanyName != nil {
		trimmed := strings.TrimSpace(*input.CompanyName)
		if trimmed == "" {
			return ApplicationDetails{}, fmt.Errorf("%w: companyName cannot be empty", ErrInvalidInput)
		}
		input.CompanyName = &trimmed
	}
	trimStringPtr(&input.PositionTitle)
	trimStringPtr(&input.SourceURL)
	trimStringPtr(&input.WorkType)
	trimStringPtr(&input.Salary)
	trimStringPtr(&input.PositionDescription)
	trimStringPtr(&input.TechStack)

	return s.repo.UpdateApplication(ctx, input, s.now())
}

func (s *Service) GetApplication(ctx context.Context, id int64) (ApplicationDetails, error) {
	if id <= 0 {
		return ApplicationDetails{}, ErrNotFound
	}
	return s.repo.GetApplication(ctx, id)
}

func (s *Service) ListApplications(ctx context.Context, input ListApplicationsInput) (ListApplicationsOutput, error) {
	input.CompanyName = strings.TrimSpace(input.CompanyName)
	input.PositionTitle = strings.TrimSpace(input.PositionTitle)
	if input.CurrentStatus != "" && !input.CurrentStatus.Valid() {
		return ListApplicationsOutput{}, ErrInvalidStatus
	}
	if input.Limit < 0 || input.Offset < 0 {
		return ListApplicationsOutput{}, fmt.Errorf("%w: limit and offset must be non-negative", ErrInvalidInput)
	}
	return s.repo.ListApplications(ctx, input)
}

func (s *Service) AddComment(ctx context.Context, input AddCommentInput) (Comment, error) {
	input.Body = strings.TrimSpace(input.Body)
	if input.ApplicationID <= 0 {
		return Comment{}, fmt.Errorf("%w: applicationId must be positive", ErrInvalidInput)
	}
	if input.Body == "" {
		return Comment{}, fmt.Errorf("%w: body is required", ErrInvalidInput)
	}
	return s.repo.AddComment(ctx, input, s.now())
}

func (s *Service) ChangeStatus(ctx context.Context, input ChangeStatusInput) (StatusChange, error) {
	input.Note = strings.TrimSpace(input.Note)
	if input.ApplicationID <= 0 {
		return StatusChange{}, fmt.Errorf("%w: applicationId must be positive", ErrInvalidInput)
	}
	if !input.Status.Valid() {
		return StatusChange{}, ErrInvalidStatus
	}
	return s.repo.ChangeStatus(ctx, input, s.now())
}

func (s *Service) UploadCV(ctx context.Context, input UploadCVInput) (Document, error) {
	return s.uploadDocument(ctx, input.ApplicationID, DocumentTypeCV, input.Filename, input.Body)
}

func (s *Service) UploadCVFromPath(ctx context.Context, input UploadCVFromPathInput) (Document, error) {
	if input.ApplicationID <= 0 {
		return Document{}, fmt.Errorf("%w: applicationId must be positive", ErrInvalidInput)
	}
	path := strings.TrimSpace(input.FilePath)
	if path == "" {
		return Document{}, fmt.Errorf("%w: filePath is required", ErrInvalidInput)
	}

	file, err := os.Open(path)
	if err != nil {
		return Document{}, err
	}
	defer file.Close()

	return s.UploadCV(ctx, UploadCVInput{
		ApplicationID: input.ApplicationID,
		Filename:      filepath.Base(path),
		Body:          file,
	})
}

func (s *Service) UploadCoverLetter(ctx context.Context, input UploadCoverLetterInput) (Document, error) {
	return s.uploadDocument(ctx, input.ApplicationID, DocumentTypeCoverLetter, input.Filename, input.Body)
}

func (s *Service) UploadCoverLetterFromPath(ctx context.Context, input UploadCoverLetterFromPathInput) (Document, error) {
	if input.ApplicationID <= 0 {
		return Document{}, fmt.Errorf("%w: applicationId must be positive", ErrInvalidInput)
	}
	path := strings.TrimSpace(input.FilePath)
	if path == "" {
		return Document{}, fmt.Errorf("%w: filePath is required", ErrInvalidInput)
	}

	file, err := os.Open(path)
	if err != nil {
		return Document{}, err
	}
	defer file.Close()

	return s.UploadCoverLetter(ctx, UploadCoverLetterInput{
		ApplicationID: input.ApplicationID,
		Filename:      filepath.Base(path),
		Body:          file,
	})
}

func trimStringPtr(value **string) {
	if *value == nil {
		return
	}
	trimmed := strings.TrimSpace(**value)
	*value = &trimmed
}

func (s *Service) uploadDocument(ctx context.Context, applicationID int64, documentType, filename string, body io.Reader) (Document, error) {
	if applicationID <= 0 {
		return Document{}, fmt.Errorf("%w: applicationId must be positive", ErrInvalidInput)
	}
	if body == nil {
		return Document{}, fmt.Errorf("%w: body is required", ErrInvalidInput)
	}
	if s.files == nil {
		return Document{}, errors.New("file store is not configured")
	}

	if _, err := s.repo.GetApplication(ctx, applicationID); err != nil {
		return Document{}, err
	}

	stored, err := s.files.SaveDocument(ctx, applicationID, documentType, strings.TrimSpace(filename), body)
	if err != nil {
		return Document{}, err
	}

	document, err := s.repo.AddDocument(ctx, AddDocumentRecordInput{
		ApplicationID:    applicationID,
		DocumentType:     documentType,
		OriginalFilename: stored.OriginalFilename,
		ContentType:      stored.ContentType,
		StoragePath:      stored.StoragePath,
		SHA256:           stored.SHA256,
		SizeBytes:        stored.SizeBytes,
	}, s.now())
	if err != nil {
		_ = s.files.Delete(ctx, stored.StoragePath)
		return Document{}, err
	}

	return document, nil
}
