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

func TestToolsCallListApplicationsReturnsDataInContent(t *testing.T) {
	t.Parallel()

	in := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize"}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_applications","arguments":{}}}`,
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

	var resp struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			StructuredContent app.ListApplicationsOutput `json:"structuredContent"`
			IsError           bool                       `json:"isError"`
		} `json:"result"`
		Error *rpcError `json:"error"`
	}

	if err := json.Unmarshal([]byte(lines[1]), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("tools/call error = %+v", resp.Error)
	}
	if resp.Result.IsError {
		t.Fatalf("tools/call returned isError=true")
	}
	if len(resp.Result.Content) != 1 {
		t.Fatalf("content items = %d, want 1", len(resp.Result.Content))
	}
	if got := resp.Result.Content[0].Text; !strings.Contains(got, `"items"`) || !strings.Contains(got, `"OpenAI"`) {
		t.Fatalf("content text = %q, want JSON payload with list data", got)
	}
	if len(resp.Result.StructuredContent.Items) != 1 {
		t.Fatalf("structured items = %d, want 1", len(resp.Result.StructuredContent.Items))
	}
	if got := resp.Result.StructuredContent.Items[0].CompanyName; got != "OpenAI" {
		t.Fatalf("companyName = %q, want %q", got, "OpenAI")
	}
}

func TestToolsCallGetApplicationStatsReturnsStructuredData(t *testing.T) {
	t.Parallel()

	in := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize"}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_application_stats","arguments":{}}}`,
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

	var resp struct {
		Result struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			StructuredContent app.ApplicationStats `json:"structuredContent"`
			IsError           bool                 `json:"isError"`
		} `json:"result"`
		Error *rpcError `json:"error"`
	}

	if err := json.Unmarshal([]byte(lines[1]), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("tools/call error = %+v", resp.Error)
	}
	if resp.Result.IsError {
		t.Fatalf("tools/call returned isError=true")
	}
	if got := resp.Result.StructuredContent.Total; got != 1 {
		t.Fatalf("total = %d, want 1", got)
	}
	if got := resp.Result.StructuredContent.ByStatus[app.StatusApplied]; got != 1 {
		t.Fatalf("applied count = %d, want 1", got)
	}
	if got := resp.Result.Content[0].Text; !strings.Contains(got, `"byStatus"`) {
		t.Fatalf("content text = %q, want JSON payload", got)
	}
}

func TestToolsCallGetApplicationTimelineReturnsEvents(t *testing.T) {
	t.Parallel()

	in := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize"}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get_application_timeline","arguments":{"applicationId":1}}}`,
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

	var resp struct {
		Result struct {
			StructuredContent app.ApplicationTimeline `json:"structuredContent"`
		} `json:"result"`
		Error *rpcError `json:"error"`
	}

	if err := json.Unmarshal([]byte(lines[1]), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("tools/call error = %+v", resp.Error)
	}
	if got := resp.Result.StructuredContent.Total; got != 3 {
		t.Fatalf("timeline total = %d, want 3", got)
	}
	if got := resp.Result.StructuredContent.Events[0].Type; got != "document" {
		t.Fatalf("first event type = %q, want %q", got, "document")
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
	now := time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC)
	return app.ApplicationDetails{
		ApplicationSummary: app.ApplicationSummary{
			ID:                  1,
			CompanyName:         "OpenAI",
			PositionTitle:       "Go Backend Engineer",
			CurrentStatus:       app.StatusApplied,
			LastStatusChangedAt: now,
			CreatedAt:           now.Add(-2 * time.Hour),
			UpdatedAt:           now,
		},
		Comments: []app.Comment{
			{ID: 1, ApplicationID: 1, Body: "sent resume", CreatedAt: now.Add(-time.Hour)},
		},
		StatusHistory: []app.StatusChange{
			{ID: 1, ApplicationID: 1, Status: app.StatusApplied, ChangedAt: now.Add(-90 * time.Minute)},
		},
		Documents: []app.Document{
			{ID: 1, ApplicationID: 1, DocumentType: app.DocumentTypeCV, OriginalFilename: "cv.pdf", UploadedAt: now},
		},
	}, nil
}
func (f *fakeRepo) ListApplications(context.Context, app.ListApplicationsInput) (app.ListApplicationsOutput, error) {
	now := time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC)
	return app.ListApplicationsOutput{
		Items: []app.ApplicationSummary{
			{
				ID:                  1,
				CompanyName:         "OpenAI",
				PositionTitle:       "Go Backend Engineer",
				CurrentStatus:       app.StatusApplied,
				LastStatusChangedAt: now,
				CreatedAt:           now,
				UpdatedAt:           now,
			},
		},
		Total: 1,
	}, nil
}
func (f *fakeRepo) SearchApplications(context.Context, app.SearchApplicationsInput) (app.ListApplicationsOutput, error) {
	return f.ListApplications(context.Background(), app.ListApplicationsInput{})
}
func (f *fakeRepo) GetRecentApplications(context.Context, app.RecentApplicationsInput) (app.ListApplicationsOutput, error) {
	return f.ListApplications(context.Background(), app.ListApplicationsInput{})
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
func (f *fakeRepo) ListDocuments(context.Context, int64) (app.DocumentList, error) {
	now := time.Date(2026, 4, 13, 12, 0, 0, 0, time.UTC)
	return app.DocumentList{
		ApplicationID: 1,
		Items: []app.Document{
			{ID: 1, ApplicationID: 1, DocumentType: app.DocumentTypeCV, OriginalFilename: "cv.pdf", UploadedAt: now},
		},
		Total: 1,
	}, nil
}
func (f *fakeRepo) GetApplicationStats(context.Context) (app.ApplicationStats, error) {
	return app.ApplicationStats{
		Total: 1,
		ByStatus: map[app.ApplicationStatus]int{
			app.StatusApplied: 1,
		},
	}, nil
}

type fakeFileStore struct{}

func (f *fakeFileStore) SaveDocument(context.Context, int64, string, string, io.Reader) (app.StoredFile, error) {
	return app.StoredFile{}, nil
}
func (f *fakeFileStore) Delete(context.Context, string) error { return nil }
