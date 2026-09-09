package repository

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"

	"github.com/amirhossein2831/chat_bot/gateway/internal/config"
	pgrepo "github.com/amirhossein2831/chat_bot/gateway/internal/repository/postgres"
	redisrepo "github.com/amirhossein2831/chat_bot/gateway/internal/repository/redis"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/auth"
)

var Module = fx.Module("repository",
	fx.Provide(
		fx.Annotate(
			func(db *pgxpool.Pool) *pgrepo.UserRepository { return pgrepo.NewUserRepository(db) },
			fx.As(new(auth.Users)),
		),
		fx.Annotate(
			func(client *redis.Client, cfg *config.Config) *redisrepo.SessionRepository {
				return redisrepo.NewSessionRepository(client, sessionTTL(cfg))
			},
			fx.As(new(auth.Sessions)),
		),
	),
)

func sessionTTL(cfg *config.Config) time.Duration { return cfg.Auth.SessionTTL() }
