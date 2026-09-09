package logger

import (
	"log/slog"
	"os"

	"github.com/amirhossein2831/chat_bot/gateway/internal/config"
)

func New(cfg *config.Config) *slog.Logger {
	level := slog.LevelDebug
	if cfg.IsProduction() {
		level = slog.LevelInfo
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
