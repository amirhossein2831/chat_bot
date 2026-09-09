package chat

import (
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/httpio"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/router"
)

func (c *Controller) Routes(r router.Routers) {
	r.Protected.Post("/chat", c.Send)
	r.Protected.Get("/chat/{id}/stream", c.Reattach)
	r.Protected.Post("/chat/{id}/stop", httpio.Wrap(c.Stop))
}
