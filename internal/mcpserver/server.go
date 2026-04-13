package mcpserver

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	app "github.com/vasary/job-search-mcp/internal/application"
)

const protocolVersion = "2025-11-25"

type Server struct {
	service *app.Service
	input   io.Reader
	output  io.Writer

	mu          sync.RWMutex
	initialized bool
}

func New(service *app.Service) *Server {
	return &Server{
		service: service,
		input:   os.Stdin,
		output:  os.Stdout,
	}
}

func (s *Server) Run(ctx context.Context) error {
	scanner := bufio.NewScanner(s.input)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var request rpcRequest
		if err := json.Unmarshal(line, &request); err != nil {
			if err := s.write(rpcResponse{
				JSONRPC: "2.0",
				Error:   &rpcError{Code: -32700, Message: "parse error"},
			}); err != nil {
				return err
			}
			continue
		}

		if err := s.handle(ctx, request); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func (s *Server) handle(ctx context.Context, request rpcRequest) error {
	switch request.Method {
	case "ping":
		return s.reply(request.ID, map[string]any{})
	case "initialize":
		s.setInitialized(false)
		return s.reply(request.ID, map[string]any{
			"protocolVersion": protocolVersion,
			"capabilities": map[string]any{
				"tools": map[string]any{
					"listChanged": false,
				},
			},
			"serverInfo": map[string]any{
				"name":    "job-search-service",
				"version": "0.2.0",
			},
		})
	case "notifications/initialized":
		s.setInitialized(true)
		return nil
	case "tools/list":
		if !s.isInitialized() {
			return s.replyError(request.ID, -32002, "server is not initialized")
		}
		return s.reply(request.ID, map[string]any{"tools": toolDefinitions()})
	case "tools/call":
		if !s.isInitialized() {
			return s.replyError(request.ID, -32002, "server is not initialized")
		}
		return s.handleToolCall(ctx, request)
	default:
		if request.ID == nil {
			return nil
		}
		return s.replyError(request.ID, -32601, "method not found")
	}
}

func (s *Server) handleToolCall(ctx context.Context, request rpcRequest) error {
	var params callToolParams
	if err := decodeParams(request.Params, &params); err != nil {
		return s.replyError(request.ID, -32602, "invalid params")
	}

	var (
		result any
		err    error
	)

	switch params.Name {
	case "create_application":
		var input app.CreateApplicationInput
		err = decodeArguments(params.Arguments, &input)
		if err == nil {
			result, err = s.service.CreateApplication(ctx, input)
		}
	case "update_application":
		var input app.UpdateApplicationInput
		err = decodeArguments(params.Arguments, &input)
		if err == nil {
			result, err = s.service.UpdateApplication(ctx, input)
		}
	case "list_applications":
		var input app.ListApplicationsInput
		err = decodeArguments(params.Arguments, &input)
		if err == nil {
			result, err = s.service.ListApplications(ctx, input)
		}
	case "get_application":
		var input getApplicationInput
		err = decodeArguments(params.Arguments, &input)
		if err == nil {
			result, err = s.service.GetApplication(ctx, input.ID)
		}
	case "add_comment":
		var input app.AddCommentInput
		err = decodeArguments(params.Arguments, &input)
		if err == nil {
			result, err = s.service.AddComment(ctx, input)
		}
	case "change_status":
		var input app.ChangeStatusInput
		err = decodeArguments(params.Arguments, &input)
		if err == nil {
			result, err = s.service.ChangeStatus(ctx, input)
		}
	case "upload_cv_from_path":
		var input app.UploadCVFromPathInput
		err = decodeArguments(params.Arguments, &input)
		if err == nil {
			result, err = s.service.UploadCVFromPath(ctx, input)
		}
	case "upload_cover_letter_from_path":
		var input app.UploadCoverLetterFromPathInput
		err = decodeArguments(params.Arguments, &input)
		if err == nil {
			result, err = s.service.UploadCoverLetterFromPath(ctx, input)
		}
	default:
		return s.replyError(request.ID, -32602, fmt.Sprintf("unknown tool %q", params.Name))
	}

	if err != nil {
		return s.replyToolError(request.ID, err)
	}

	return s.reply(request.ID, map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": "ok"},
		},
		"structuredContent": result,
		"isError":           false,
	})
}

func (s *Server) replyToolError(id any, err error) error {
	message := err.Error()
	switch {
	case errors.Is(err, app.ErrNotFound):
		message = "application not found"
	case errors.Is(err, app.ErrInvalidStatus):
		message = "invalid status, allowed values: applied, screening, interview, offer, rejected, withdrawn, accepted"
	case errors.Is(err, app.ErrInvalidPDF):
		message = "file must be a PDF"
	}

	return s.reply(id, map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": message},
		},
		"isError": true,
	})
}

func (s *Server) reply(id any, result any) error {
	return s.write(rpcResponse{JSONRPC: "2.0", ID: id, Result: result})
}

func (s *Server) replyError(id any, code int, message string) error {
	return s.write(rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: message},
	})
}

func (s *Server) write(response rpcResponse) error {
	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(s.output, "%s\n", payload)
	return err
}

func (s *Server) setInitialized(value bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.initialized = value
}

func (s *Server) isInitialized() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.initialized
}

func decodeParams(raw json.RawMessage, out any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

func decodeArguments(raw json.RawMessage, out any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type callToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

type getApplicationInput struct {
	ID int64 `json:"id"`
}

func toolDefinitions() []map[string]any {
	return []map[string]any{
		{
			"name":        "create_application",
			"description": "Create a new job application with an initial status and optional first comment",
			"inputSchema": map[string]any{
				"type":     "object",
				"required": []string{"companyName"},
				"properties": map[string]any{
					"companyName":         map[string]any{"type": "string", "minLength": 1},
					"positionTitle":       map[string]any{"type": "string"},
					"sourceUrl":           map[string]any{"type": "string"},
					"workType":            map[string]any{"type": "string"},
					"salary":              map[string]any{"type": "string"},
					"positionDescription": map[string]any{"type": "string"},
					"techStack":           map[string]any{"type": "string"},
					"initialStatus":       statusSchema(),
					"initialStatusNote":   map[string]any{"type": "string"},
					"initialComment":      map[string]any{"type": "string"},
				},
			},
		},
		{
			"name":        "update_application",
			"description": "Update the editable fields of an existing job application",
			"inputSchema": map[string]any{
				"type":     "object",
				"required": []string{"id"},
				"properties": map[string]any{
					"id":                  map[string]any{"type": "integer", "minimum": 1},
					"companyName":         map[string]any{"type": "string"},
					"positionTitle":       map[string]any{"type": "string"},
					"sourceUrl":           map[string]any{"type": "string"},
					"workType":            map[string]any{"type": "string"},
					"salary":              map[string]any{"type": "string"},
					"positionDescription": map[string]any{"type": "string"},
					"techStack":           map[string]any{"type": "string"},
				},
			},
		},
		{
			"name":        "list_applications",
			"description": "List job applications with filters by company, position title and current status",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"companyName":   map[string]any{"type": "string"},
					"positionTitle": map[string]any{"type": "string"},
					"currentStatus": statusSchema(),
					"limit":         map[string]any{"type": "integer", "minimum": 0},
					"offset":        map[string]any{"type": "integer", "minimum": 0},
				},
			},
		},
		{
			"name":        "get_application",
			"description": "Get one application with comments, status history and documents",
			"inputSchema": map[string]any{
				"type":     "object",
				"required": []string{"id"},
				"properties": map[string]any{
					"id": map[string]any{"type": "integer", "minimum": 1},
				},
			},
		},
		{
			"name":        "add_comment",
			"description": "Add a timestamped comment to an application",
			"inputSchema": map[string]any{
				"type":     "object",
				"required": []string{"applicationId", "body"},
				"properties": map[string]any{
					"applicationId": map[string]any{"type": "integer", "minimum": 1},
					"body":          map[string]any{"type": "string", "minLength": 1},
				},
			},
		},
		{
			"name":        "change_status",
			"description": "Add a new status transition entry for an application",
			"inputSchema": map[string]any{
				"type":     "object",
				"required": []string{"applicationId", "status"},
				"properties": map[string]any{
					"applicationId": map[string]any{"type": "integer", "minimum": 1},
					"status":        statusSchema(),
					"note":          map[string]any{"type": "string"},
				},
			},
		},
		{
			"name":        "upload_cv_from_path",
			"description": "Upload a CV PDF from a local file path and attach it to an application",
			"inputSchema": map[string]any{
				"type":     "object",
				"required": []string{"applicationId", "filePath"},
				"properties": map[string]any{
					"applicationId": map[string]any{"type": "integer", "minimum": 1},
					"filePath":      map[string]any{"type": "string", "minLength": 1},
				},
			},
		},
		{
			"name":        "upload_cover_letter_from_path",
			"description": "Upload a cover letter PDF from a local file path and attach it to an application",
			"inputSchema": map[string]any{
				"type":     "object",
				"required": []string{"applicationId", "filePath"},
				"properties": map[string]any{
					"applicationId": map[string]any{"type": "integer", "minimum": 1},
					"filePath":      map[string]any{"type": "string", "minLength": 1},
				},
			},
		},
	}
}

func statusSchema() map[string]any {
	return map[string]any{
		"type": "string",
		"enum": []string{"applied", "screening", "interview", "offer", "rejected", "withdrawn", "accepted"},
	}
}
