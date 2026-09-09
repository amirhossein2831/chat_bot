package health

import "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/router"

func (c *Controller) Routes(r router.Routers) {
	r.Root.Get("/healthz", c.Live)
}
