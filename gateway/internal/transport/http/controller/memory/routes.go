package memory

import (
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/httpio"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/router"
)

func (c *Controller) Routes(r router.Routers) {
	r.Protected.Get("/settings", httpio.Wrap(c.Settings))
	r.Protected.Put("/settings", httpio.Wrap(c.SetSettings))

	r.Protected.Get("/memory", httpio.Wrap(c.List))
	r.Protected.Post("/memory", httpio.Wrap(c.Add))
	r.Protected.Delete("/memory/{id}", httpio.Wrap(c.Delete))
}
