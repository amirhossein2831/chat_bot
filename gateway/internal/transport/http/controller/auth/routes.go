package auth

import (
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/httpio"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/router"
)

func (c *Controller) Routes(r router.Routers) {
	r.Public.Post("/auth/register", httpio.Wrap(c.Register))
	r.Public.Post("/auth/login", httpio.Wrap(c.Login))

	r.Protected.Post("/auth/logout", httpio.Wrap(c.Logout))
	r.Protected.Get("/me", httpio.Wrap(c.Me))
}
