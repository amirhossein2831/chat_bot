package http

import "go.uber.org/fx"

var Module = fx.Module("transport",
	fx.Provide(NewServer),
)
