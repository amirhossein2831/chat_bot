package auth

import "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/router"

func (c *Controller) Routes(r router.Routers) {
	r.Public.Post("/auth/register", c.Register)
	r.Public.Post("/auth/login", c.Login)

	r.Protected.Post("/auth/logout", c.Logout)
	r.Protected.Get("/me", c.Me)
}
