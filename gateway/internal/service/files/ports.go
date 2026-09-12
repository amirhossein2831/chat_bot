package files

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/amirhossein2831/chat_bot/gateway/internal/model"
)

type Repository interface {
	Create(ctx context.Context, a model.Attachment) error
	Get(ctx context.Context, id uuid.UUID) (model.Attachment, error)
	GetOwned(ctx context.Context, id, userID uuid.UUID) (model.Attachment, error)
	SetStatus(ctx context.Context, id uuid.UUID, status model.ExtractionStatus, textKey, failure *string) error
	CountToday(ctx context.Context, userID uuid.UUID) (int, error)
	Attach(ctx context.Context, ids []uuid.UUID, messageID, userID uuid.UUID) error
	OfMessage(ctx context.Context, messageID uuid.UUID) ([]model.Attachment, error)
	OwnedByIDs(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) ([]model.Attachment, error)
}

type ObjectStore interface {
	PresignPut(ctx context.Context, key string, ttl time.Duration) (string, error)
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error
	Stat(ctx context.Context, key string) (int64, error)
}

type Queue interface {
	Enqueue(ctx context.Context, attachmentID uuid.UUID) error
}

// Pages reads a PDF. Declared here because this service is what needs it; the
// pdfium adapter satisfies it without knowing this package exists.
type Pages interface {
	Pages(document []byte) ([]string, error)
	Render(document []byte, dpi int, limit int) ([][]byte, error)
}

// Recognizer is the OCR service. Absent when it is not deployed, in which case
// scans are reported as unsupported rather than failing.
type Recognizer interface {
	Recognize(ctx context.Context, pages [][]byte, languages []string) (string, error)
}
