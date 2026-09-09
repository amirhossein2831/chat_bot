package infrastructure

import (
	"go.uber.org/fx"

	"github.com/amirhossein2831/chat_bot/gateway/internal/infrastructure/postgres"
	"github.com/amirhossein2831/chat_bot/gateway/internal/logger"
)

var Module = fx.Module("infrastructure",
	fx.Provide(
		logger.New,
		postgres.NewPool,
	),
)
