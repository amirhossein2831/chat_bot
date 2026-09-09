package main

import (
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/amirhossein2831/chat_bot/gateway/internal/config"
	"github.com/amirhossein2831/chat_bot/gateway/internal/infrastructure"
	"github.com/amirhossein2831/chat_bot/gateway/internal/repository"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service"
	httptransport "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/controller"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/router"
)

func options() fx.Option {
	return fx.Options(
		config.Module,
		infrastructure.Module,
		repository.Module,
		service.Module,
		controller.Module,
		router.Module,
		httptransport.Module,

		// fx constructs lazily; without this nothing is built and a broken
		// config or unreachable dependency never surfaces.
		fx.Invoke(func(*http.Server) {}),

		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)
}

func main() {
	fx.New(options()).Run()
}
