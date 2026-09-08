// Command gateway is the public API surface of the platform.
package main

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/amirhossein2831/chat_bot/gateway/internal/config"
	"github.com/amirhossein2831/chat_bot/gateway/internal/infrastructure/postgres"
)

// options builds the application graph.
//
// It is a function rather than a variable so that the graph validated in tests
// is byte-for-byte the graph main runs. Every module registers itself here.
func options() fx.Option {
	return fx.Options(
		config.Module,
		postgres.Module,

		// fx constructs lazily: without an invocation nothing is built, and a
		// broken configuration or an unreachable database would never surface.
		// These force construction so failures happen at boot, loudly.
		fx.Invoke(func(*config.Config) {}),
		fx.Invoke(func(*pgxpool.Pool) {}),
	)
}

func main() {
	fx.New(options()).Run()
}
