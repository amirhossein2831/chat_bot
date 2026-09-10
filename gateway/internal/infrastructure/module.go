package infrastructure

import (
	"context"

	"go.uber.org/fx"

	"github.com/amirhossein2831/chat_bot/gateway/internal/infrastructure/objectstore"
	"github.com/amirhossein2831/chat_bot/gateway/internal/infrastructure/pdf"
	"github.com/amirhossein2831/chat_bot/gateway/internal/infrastructure/postgres"
	"github.com/amirhossein2831/chat_bot/gateway/internal/infrastructure/redis"
	"github.com/amirhossein2831/chat_bot/gateway/internal/logger"
)

var Module = fx.Module("infrastructure",
	fx.Provide(
		logger.New,
		postgres.NewPool,
		redis.NewClient,
		objectstore.New,
		newPDFReader,
	),
)

// The pdfium runtime is a WebAssembly module compiled once and held for the
// life of the process, so it is built here rather than per document.
func newPDFReader(lc fx.Lifecycle) (*pdf.Reader, error) {
	reader, err := pdf.New()
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error { return reader.Close() },
	})
	return reader, nil
}
