package mapper

import (
	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/dto"
)

func ToAttachmentResponse(a model.Attachment) dto.AttachmentResponse {
	return dto.AttachmentResponse{
		ID:       a.ID.String(),
		Filename: a.Filename,
		MIME:     a.MIME,
		Size:     a.SizeBytes,
		Status:   string(a.ExtractionStatus),
		Error:    a.ExtractionError,
	}
}
