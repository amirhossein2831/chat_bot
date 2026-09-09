package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/amirhossein2831/chat_bot/gateway/internal/config"
	"github.com/amirhossein2831/chat_bot/gateway/internal/infrastructure/postgres"
)

func options() fx.Option {
	return fx.Options(
		config.Module,
		postgres.Module,

		// fx constructs lazily; without these nothing is built and a broken
		// config or unreachable database never surfaces.
		fx.Invoke(func(*config.Config) {}),
		fx.Invoke(func(*pgxpool.Pool) {}),
	)
}

func main() {
	fx.New(options()).Run()
}
