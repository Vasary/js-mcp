package mcpserver

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
