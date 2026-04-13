package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	app "github.com/vasary/job-search-mcp/internal/application"
)

func TestStdioInitializeAndToolsList(t *testing.T) {
	t.Parallel()

	in := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize"}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
	}, "\n") + "\n"

	var out bytes.Buffer
	server := New(app.NewService(&fakeRepo{}, &fakeFileStore{}))
	server.input = strings.NewReader(in)
	server.output = &out

	if err := server.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("responses = %d, want 2", len(lines))
	}

	var resp rpcResponse
	if err := json.Unmarshal([]byte(lines[1]), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("tools/list error = %+v", resp.Error)
	}
}

type fakeRepo struct{}

func (f *fakeRepo) Ping(context.Context) error { return nil }
func (f *fakeRepo) CreateApplication(context.Context, app.CreateApplicationInput, time.Time) (app.ApplicationDetails, error) {
	return app.ApplicationDetails{}, nil
}
func (f *fakeRepo) UpdateApplication(context.Context, app.UpdateApplicationInput, time.Time) (app.ApplicationDetails, error) {
	return app.ApplicationDetails{}, nil
}
func (f *fakeRepo) GetApplication(context.Context, int64) (app.ApplicationDetails, error) {
	return app.ApplicationDetails{}, nil
}
func (f *fakeRepo) ListApplications(context.Context, app.ListApplicationsInput) (app.ListApplicationsOutput, error) {
	return app.ListApplicationsOutput{}, nil
}
func (f *fakeRepo) AddComment(context.Context, app.AddCommentInput, time.Time) (app.Comment, error) {
	return app.Comment{}, nil
}
func (f *fakeRepo) ChangeStatus(context.Context, app.ChangeStatusInput, time.Time) (app.StatusChange, error) {
	return app.StatusChange{}, nil
}
func (f *fakeRepo) AddDocument(context.Context, app.AddDocumentRecordInput, time.Time) (app.Document, error) {
	return app.Document{}, nil
}

type fakeFileStore struct{}

func (f *fakeFileStore) SaveDocument(context.Context, int64, string, string, io.Reader) (app.StoredFile, error) {
	return app.StoredFile{}, nil
}
func (f *fakeFileStore) Delete(context.Context, string) error { return nil }
