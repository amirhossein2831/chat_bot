package infrastructure

import (
	"go.uber.org/fx"

	"github.com/amirhossein2831/chat_bot/gateway/internal/infrastructure/postgres"
)

var Module = fx.Module("infrastructure",
	fx.Provide(postgres.NewPool),
)
