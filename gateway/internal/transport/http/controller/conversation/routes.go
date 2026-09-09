package conversation

import (
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/httpio"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/router"
)

func (c *Controller) Routes(r router.Routers) {
	r.Protected.Get("/conversations", httpio.Wrap(c.List))
	r.Protected.Post("/conversations", httpio.Wrap(c.Create))
	r.Protected.Get("/conversations/{id}/messages", httpio.Wrap(c.Messages))
	r.Protected.Patch("/conversations/{id}", httpio.Wrap(c.Rename))
	r.Protected.Delete("/conversations/{id}", httpio.Wrap(c.Archive))
	r.Protected.Post("/conversations/{id}/branch", httpio.Wrap(c.SwitchBranch))
}
