package application_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/vasary/job-search-mcp/internal/application"
)

func TestCreateApplication_DefaultsStatus(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{}
	svc := application.NewService(repo, nil)

	_, err := svc.CreateApplication(context.Background(), application.CreateApplicationInput{
		PositionTitle: "Go Backend Engineer",
	})
	if err != nil {
		t.Fatalf("CreateApplication() error = %v", err)
	}

	if repo.created.InitialStatus != application.StatusApplied {
		t.Fatalf("initial status = %q, want %q", repo.created.InitialStatus, application.StatusApplied)
	}
}

func TestCreateApplication_RequiresPositionTitle(t *testing.T) {
	t.Parallel()

	svc := application.NewService(&fakeRepo{}, nil)
	_, err := svc.CreateApplication(context.Background(), application.CreateApplicationInput{
		CompanyName: "OpenAI",
	})
	if !errors.Is(err, application.ErrInvalidInput) {
		t.Fatalf("CreateApplication() error = %v, want ErrInvalidInput", err)
	}
}

func TestUpdateApplication_AllowsClearingCompanyName(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{}
	svc := application.NewService(repo, nil)
	empty := "   "

	_, err := svc.UpdateApplication(context.Background(), application.UpdateApplicationInput{
		ID:          1,
		CompanyName: &empty,
	})
	if err != nil {
		t.Fatalf("UpdateApplication() error = %v", err)
	}
	if repo.updated.CompanyName == nil || *repo.updated.CompanyName != "" {
		t.Fatalf("companyName = %#v, want cleared empty string pointer", repo.updated.CompanyName)
	}
}

func TestChangeStatus_ValidatesStatus(t *testing.T) {
	t.Parallel()

	svc := application.NewService(&fakeRepo{}, nil)
	_, err := svc.ChangeStatus(context.Background(), application.ChangeStatusInput{
		ApplicationID: 1,
		Status:        application.ApplicationStatus("bad"),
	})
	if !errors.Is(err, application.ErrInvalidStatus) {
		t.Fatalf("ChangeStatus() error = %v, want ErrInvalidStatus", err)
	}
}

func TestUploadCVFromPath_DelegatesToFileStore(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{
		getResult: application.ApplicationDetails{
			ApplicationSummary: application.ApplicationSummary{ID: 7},
		},
		documentResult: application.Document{ID: 11},
	}
	files := &fakeFileStore{
		saveResult: application.StoredFile{
			OriginalFilename: "resume.pdf",
			ContentType:      "application/pdf",
			StoragePath:      "/tmp/resume.pdf",
			SHA256:           "abc",
			SizeBytes:        42,
		},
	}

	svc := application.NewService(repo, files)
	doc, err := svc.UploadCV(context.Background(), application.UploadCVInput{
		ApplicationID: 7,
		Filename:      "resume.pdf",
		Body:          bytes.NewBufferString("%PDF-test"),
	})
	if err != nil {
		t.Fatalf("UploadCV() error = %v", err)
	}
	if doc.ID != 11 {
		t.Fatalf("UploadCV() document ID = %d, want 11", doc.ID)
	}
	if repo.addedDocument.DocumentType != application.DocumentTypeCV {
		t.Fatalf("UploadCV() document type = %q", repo.addedDocument.DocumentType)
	}
}

func TestUploadCoverLetter_DelegatesToFileStore(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{
		getResult: application.ApplicationDetails{
			ApplicationSummary: application.ApplicationSummary{ID: 7},
		},
		documentResult: application.Document{ID: 12},
	}
	files := &fakeFileStore{
		saveResult: application.StoredFile{
			OriginalFilename: "cover-letter.pdf",
			ContentType:      "application/pdf",
			StoragePath:      "/tmp/cover-letter.pdf",
			SHA256:           "def",
			SizeBytes:        55,
		},
	}

	svc := application.NewService(repo, files)
	doc, err := svc.UploadCoverLetter(context.Background(), application.UploadCoverLetterInput{
		ApplicationID: 7,
		Filename:      "cover-letter.pdf",
		Body:          bytes.NewBufferString("%PDF-test"),
	})
	if err != nil {
		t.Fatalf("UploadCoverLetter() error = %v", err)
	}
	if doc.ID != 12 {
		t.Fatalf("UploadCoverLetter() document ID = %d, want 12", doc.ID)
	}
	if repo.addedDocument.DocumentType != application.DocumentTypeCoverLetter {
		t.Fatalf("UploadCoverLetter() document type = %q", repo.addedDocument.DocumentType)
	}
}

type fakeRepo struct {
	created        application.CreateApplicationInput
	updated        application.UpdateApplicationInput
	getResult      application.ApplicationDetails
	documentResult application.Document
	addedDocument  application.AddDocumentRecordInput
}

func (f *fakeRepo) Ping(context.Context) error { return nil }

func (f *fakeRepo) CreateApplication(_ context.Context, input application.CreateApplicationInput, _ time.Time) (application.ApplicationDetails, error) {
	f.created = input
	return application.ApplicationDetails{}, nil
}

func (f *fakeRepo) UpdateApplication(_ context.Context, input application.UpdateApplicationInput, _ time.Time) (application.ApplicationDetails, error) {
	f.updated = input
	return application.ApplicationDetails{}, nil
}

func (f *fakeRepo) GetApplication(context.Context, int64) (application.ApplicationDetails, error) {
	return f.getResult, nil
}

func (f *fakeRepo) ListApplications(context.Context, application.ListApplicationsInput) (application.ListApplicationsOutput, error) {
	return application.ListApplicationsOutput{}, nil
}

func (f *fakeRepo) SearchApplications(context.Context, application.SearchApplicationsInput) (application.ListApplicationsOutput, error) {
	return application.ListApplicationsOutput{}, nil
}

func (f *fakeRepo) GetRecentApplications(context.Context, application.RecentApplicationsInput) (application.ListApplicationsOutput, error) {
	return application.ListApplicationsOutput{}, nil
}

func (f *fakeRepo) AddComment(context.Context, application.AddCommentInput, time.Time) (application.Comment, error) {
	return application.Comment{}, nil
}

func (f *fakeRepo) ChangeStatus(context.Context, application.ChangeStatusInput, time.Time) (application.StatusChange, error) {
	return application.StatusChange{}, nil
}

func (f *fakeRepo) AddDocument(_ context.Context, input application.AddDocumentRecordInput, _ time.Time) (application.Document, error) {
	f.addedDocument = input
	return f.documentResult, nil
}

func (f *fakeRepo) ListDocuments(context.Context, int64) (application.DocumentList, error) {
	return application.DocumentList{}, nil
}

func (f *fakeRepo) GetApplicationStats(context.Context) (application.ApplicationStats, error) {
	return application.ApplicationStats{}, nil
}

type fakeFileStore struct {
	saveResult application.StoredFile
}

func (f *fakeFileStore) SaveDocument(_ context.Context, _ int64, _ string, _ string, _ io.Reader) (application.StoredFile, error) {
	return f.saveResult, nil
}

func (f *fakeFileStore) Delete(context.Context, string) error { return nil }
