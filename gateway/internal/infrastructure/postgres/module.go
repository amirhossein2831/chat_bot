package postgres

import "go.uber.org/fx"

var Module = fx.Module("infrastructure.postgres", fx.Provide(NewPool))
