package main

import (
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/amirhossein2831/chat_bot/gateway/internal/client"
	"github.com/amirhossein2831/chat_bot/gateway/internal/config"
	"github.com/amirhossein2831/chat_bot/gateway/internal/infrastructure"
	"github.com/amirhossein2831/chat_bot/gateway/internal/jobs"
	"github.com/amirhossein2831/chat_bot/gateway/internal/observability"
	"github.com/amirhossein2831/chat_bot/gateway/internal/repository"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport"
)

func options() fx.Option {
	return fx.Options(
		config.Module,
		observability.Module,
		infrastructure.Module,
		client.Module,
		jobs.Module,
		repository.Module,
		service.Module,
		transport.Module,

		// fx constructs lazily; without this nothing is built and a broken
		// config or unreachable dependency never surfaces.
		fx.Invoke(func(*http.Server) {}),
		fx.Invoke(func(jobs.WorkerClient) {}),

		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			// fx logs every constructor at info, which buries real errors.
			return &fxevent.ZapLogger{Logger: log.WithOptions(zap.IncreaseLevel(zap.WarnLevel))}
		}),
	)
}

func main() {
	fx.New(options()).Run()
}
