package dto

type PresignRequest struct {
	Filename string `json:"filename" validate:"required,max=255"`
	MIME     string `json:"mime"     validate:"required,max=128"`
	Size     int64  `json:"size"     validate:"required,min=1"`
}

type PresignResponse struct {
	AttachmentID string `json:"attachment_id"`
	URL          string `json:"url"`
	ExpiresAt    string `json:"expires_at"`
}

type AttachmentResponse struct {
	ID       string  `json:"id"`
	Filename string  `json:"filename"`
	MIME     string  `json:"mime"`
	Size     int64   `json:"size"`
	Status   string  `json:"status"`
	Error    *string `json:"error,omitempty"`
}
