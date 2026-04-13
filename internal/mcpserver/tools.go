package mcpserver

func toolDefinitions() []map[string]any {
	return []map[string]any{
		{
			"name":        "create_job_application",
			"description": "Create a tracked job application record for a vacancy you submitted to an employer",
			"inputSchema": map[string]any{
				"type":     "object",
				"required": []string{"positionTitle"},
				"properties": map[string]any{
					"companyName":         map[string]any{"type": "string"},
					"positionTitle":       map[string]any{"type": "string", "minLength": 1},
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
			"name":        "update_job_application",
			"description": "Update the editable details of a tracked job application you sent to an employer",
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
			"name":        "list_job_applications",
			"description": "List tracked job applications you have submitted to employers, with optional filters",
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
			"name":        "search_job_applications",
			"description": "Search tracked job applications by company name, position title, or tech stack",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query":         map[string]any{"type": "string"},
					"currentStatus": statusSchema(),
					"limit":         map[string]any{"type": "integer", "minimum": 0},
					"offset":        map[string]any{"type": "integer", "minimum": 0},
				},
			},
		},
		{
			"name":        "list_recent_job_applications",
			"description": "List the most recently created tracked job applications",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"limit":  map[string]any{"type": "integer", "minimum": 0},
					"offset": map[string]any{"type": "integer", "minimum": 0},
				},
			},
		},
		{
			"name":        "get_job_application",
			"description": "Get one tracked job application with its notes, status history, and attached documents",
			"inputSchema": map[string]any{
				"type":     "object",
				"required": []string{"id"},
				"properties": map[string]any{
					"id": map[string]any{"type": "integer", "minimum": 1},
				},
			},
		},
		{
			"name":        "get_job_application_timeline",
			"description": "Get one tracked job application as a unified timeline of status changes, notes, and documents",
			"inputSchema": map[string]any{
				"type":     "object",
				"required": []string{"applicationId"},
				"properties": map[string]any{
					"applicationId": map[string]any{"type": "integer", "minimum": 1},
				},
			},
		},
		{
			"name":        "list_job_application_documents",
			"description": "List the documents attached to one tracked job application",
			"inputSchema": map[string]any{
				"type":     "object",
				"required": []string{"applicationId"},
				"properties": map[string]any{
					"applicationId": map[string]any{"type": "integer", "minimum": 1},
				},
			},
		},
		{
			"name":        "get_job_application_stats",
			"description": "Get overall counts for tracked job applications grouped by current hiring status",
			"inputSchema": map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			"name":        "add_job_application_note",
			"description": "Add a timestamped note about recruiter feedback, follow-up, or interview progress",
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
			"name":        "change_job_application_status",
			"description": "Change the current hiring stage of a tracked job application and record it in history",
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
			"name":        "attach_cv_to_job_application",
			"description": "Attach a CV PDF from a local file path to a tracked job application",
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
			"name":        "attach_cover_letter_to_job_application",
			"description": "Attach a cover letter PDF from a local file path to a tracked job application",
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
