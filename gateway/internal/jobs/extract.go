package jobs

import (
	"context"

	"github.com/google/uuid"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type ExtractArgs struct {
	AttachmentID uuid.UUID `json:"attachment_id"`
}

func (ExtractArgs) Kind() string { return "extract_attachment" }

type Extractor interface {
	Extract(ctx context.Context, attachmentID uuid.UUID) error
}

type ExtractWorker struct {
	river.WorkerDefaults[ExtractArgs]

	extractor Extractor
	log       *zap.Logger
}

func NewExtractWorker(extractor Extractor, log *zap.Logger) *ExtractWorker {
	return &ExtractWorker{extractor: extractor, log: log}
}

func (w *ExtractWorker) Work(ctx context.Context, job *river.Job[ExtractArgs]) error {
	if err := w.extractor.Extract(ctx, job.Args.AttachmentID); err != nil {
		w.log.Warn("extraction failed",
			zap.String("attachment_id", job.Args.AttachmentID.String()),
			zap.Error(err))
		return err
	}
	return nil
}
