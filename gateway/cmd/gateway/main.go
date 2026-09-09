package main

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/amirhossein2831/chat_bot/gateway/internal/config"
	"github.com/amirhossein2831/chat_bot/gateway/internal/infrastructure"
	httptransport "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http"
)

func options() fx.Option {
	return fx.Options(
		config.Module,
		infrastructure.Module,
		httptransport.Module,

		// fx constructs lazily; without these nothing is built and a broken
		// config or unreachable database never surfaces. The pool invoke goes
		// away once repositories depend on it.
		fx.Invoke(func(*http.Server) {}),
		fx.Invoke(func(*pgxpool.Pool) {}),

		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)
}

func main() {
	fx.New(options()).Run()
}
