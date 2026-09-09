package main

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

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
	)
}

func main() {
	fx.New(options()).Run()
}
