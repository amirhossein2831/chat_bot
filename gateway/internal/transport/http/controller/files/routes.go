package files

import (
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/httpio"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/router"
)

func (c *Controller) Routes(r router.Routers) {
	r.Protected.Post("/files/presign", httpio.Wrap(c.Presign))
	r.Protected.Post("/files/{id}/confirm", httpio.Wrap(c.Confirm))
	r.Protected.Get("/files/{id}", httpio.Wrap(c.Get))
}
