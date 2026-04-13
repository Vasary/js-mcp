package application

import "time"

type ApplicationStatus string

const (
	StatusApplied           ApplicationStatus = "applied"
	StatusScreening         ApplicationStatus = "screening"
	StatusInterview         ApplicationStatus = "interview"
	StatusOffer             ApplicationStatus = "offer"
	StatusRejected          ApplicationStatus = "rejected"
	StatusWithdrawn         ApplicationStatus = "withdrawn"
	StatusAccepted          ApplicationStatus = "accepted"
	DocumentTypeCV          string            = "cv"
	DocumentTypeCoverLetter string            = "cover_letter"
	DefaultUploadSize       int64             = 10 << 20
)

var validStatuses = map[ApplicationStatus]struct{}{
	StatusApplied:   {},
	StatusScreening: {},
	StatusInterview: {},
	StatusOffer:     {},
	StatusRejected:  {},
	StatusWithdrawn: {},
	StatusAccepted:  {},
}

func (s ApplicationStatus) Valid() bool {
	_, ok := validStatuses[s]
	return ok
}

type ApplicationSummary struct {
	ID                  int64             `json:"id"`
	CompanyName         string            `json:"companyName"`
	PositionTitle       string            `json:"positionTitle,omitempty"`
	SourceURL           string            `json:"sourceUrl,omitempty"`
	WorkType            string            `json:"workType,omitempty"`
	Salary              string            `json:"salary,omitempty"`
	PositionDescription string            `json:"positionDescription,omitempty"`
	TechStack           string            `json:"techStack,omitempty"`
	CurrentStatus       ApplicationStatus `json:"currentStatus"`
	LastStatusChangedAt time.Time         `json:"lastStatusChangedAt"`
	CreatedAt           time.Time         `json:"createdAt"`
	UpdatedAt           time.Time         `json:"updatedAt"`
}

type ApplicationDetails struct {
	ApplicationSummary
	Comments      []Comment      `json:"comments"`
	StatusHistory []StatusChange `json:"statusHistory"`
	Documents     []Document     `json:"documents"`
}

type Comment struct {
	ID            int64     `json:"id"`
	ApplicationID int64     `json:"applicationId"`
	Body          string    `json:"body"`
	CreatedAt     time.Time `json:"createdAt"`
}

type StatusChange struct {
	ID            int64             `json:"id"`
	ApplicationID int64             `json:"applicationId"`
	Status        ApplicationStatus `json:"status"`
	Note          string            `json:"note,omitempty"`
	ChangedAt     time.Time         `json:"changedAt"`
}

type Document struct {
	ID               int64     `json:"id"`
	ApplicationID    int64     `json:"applicationId"`
	DocumentType     string    `json:"documentType"`
	OriginalFilename string    `json:"originalFilename"`
	ContentType      string    `json:"contentType"`
	StoragePath      string    `json:"storagePath"`
	SHA256           string    `json:"sha256"`
	SizeBytes        int64     `json:"sizeBytes"`
	UploadedAt       time.Time `json:"uploadedAt"`
}
