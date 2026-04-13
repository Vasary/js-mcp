package mcpserver

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	app "github.com/vasary/job-search-mcp/internal/application"
)

const protocolVersion = "2025-06-18"

type Server struct {
	service *app.Service
	input   io.Reader
	output  io.Writer
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

	initialized := false

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

		response, nextInitialized, accepted, err := s.handle(ctx, request, initialized)
		if err != nil {
			return err
		}
		initialized = nextInitialized
		if accepted {
			continue
		}
		if err := s.write(response); err != nil {
			return err
		}
	}

	return scanner.Err()
}

func (s *Server) handle(ctx context.Context, request rpcRequest, initialized bool) (rpcResponse, bool, bool, error) {
	switch request.Method {
	case "ping":
		return rpcResponse{JSONRPC: "2.0", ID: request.ID, Result: map[string]any{}}, initialized, false, nil
	case "initialize":
		return rpcResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Result: map[string]any{
				"protocolVersion": protocolVersion,
				"capabilities": map[string]any{
					"tools": map[string]any{"listChanged": false},
				},
				"serverInfo": map[string]any{
					"name":    "job-search-mcp",
					"version": "0.3.0",
				},
			},
		}, false, false, nil
	case "notifications/initialized":
		return rpcResponse{}, true, true, nil
	case "tools/list":
		if !initialized {
			return sessionErrorResponse(request.ID, "server is not initialized"), initialized, false, nil
		}
		return rpcResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Result:  map[string]any{"tools": toolDefinitions()},
		}, initialized, false, nil
	case "tools/call":
		if !initialized {
			return sessionErrorResponse(request.ID, "server is not initialized"), initialized, false, nil
		}
		response, err := s.handleToolCall(ctx, request)
		return response, initialized, false, err
	default:
		if request.ID == nil {
			return rpcResponse{}, initialized, true, nil
		}
		return rpcResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error:   &rpcError{Code: -32601, Message: "method not found"},
		}, initialized, false, nil
	}
}

func (s *Server) handleToolCall(ctx context.Context, request rpcRequest) (rpcResponse, error) {
	var params callToolParams
	if err := decodeParams(request.Params, &params); err != nil {
		return rpcResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error:   &rpcError{Code: -32602, Message: "invalid params"},
		}, nil
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
		return rpcResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
			Error:   &rpcError{Code: -32602, Message: fmt.Sprintf("unknown tool %q", params.Name)},
		}, nil
	}

	if err != nil {
		return toolErrorResponse(request.ID, err), nil
	}

	return rpcResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: map[string]any{
			"content":           []map[string]any{{"type": "text", "text": "ok"}},
			"structuredContent": result,
			"isError":           false,
		},
	}, nil
}

func toolErrorResponse(id any, err error) rpcResponse {
	message := err.Error()
	switch {
	case errors.Is(err, app.ErrNotFound):
		message = "application not found"
	case errors.Is(err, app.ErrInvalidStatus):
		message = "invalid status, allowed values: applied, screening, interview, offer, rejected, withdrawn, accepted"
	case errors.Is(err, app.ErrInvalidPDF):
		message = "file must be a PDF"
	}

	return rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result: map[string]any{
			"content": []map[string]any{{"type": "text", "text": message}},
			"isError": true,
		},
	}
}

func sessionErrorResponse(id any, message string) rpcResponse {
	return rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: -32002, Message: message},
	}
}

func (s *Server) write(response rpcResponse) error {
	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(s.output, "%s\n", payload)
	return err
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
