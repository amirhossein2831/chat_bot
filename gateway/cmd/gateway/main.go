// Command gateway is the public API surface of the platform.
package main

import (
	"go.uber.org/fx"

	"github.com/amirhossein2831/chat_bot/gateway/internal/config"
)

// options builds the application graph.
//
// It is a function rather than a variable so that the graph validated in tests
// is byte-for-byte the graph main runs. Every module registers itself here.
func options() fx.Option {
	return fx.Options(
		config.Module,

		// fx constructs lazily: without an invocation nothing is built and a
		// broken configuration would never surface. This forces config to load
		// at startup so misconfiguration fails immediately and loudly.
		fx.Invoke(func(*config.Config) {}),
	)
}

func main() {
	fx.New(options()).Run()
}
