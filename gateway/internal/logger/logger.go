package logger

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/amirhossein2831/chat_bot/gateway/internal/config"
)

func New(lc fx.Lifecycle, cfg *config.Config) (*zap.Logger, error) {
	zc := zap.NewDevelopmentConfig()
	if cfg.IsProduction() {
		zc = zap.NewProductionConfig()
		zc.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	log, err := zc.Build()
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(context.Context) error {
			_ = log.Sync()
			return nil
		},
	})

	return log, nil
}
