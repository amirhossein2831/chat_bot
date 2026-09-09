package prompt

import (
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/httpio"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/router"
)

func (c *Controller) Routes(r router.Routers) {
	r.Protected.Get("/prompts", httpio.Wrap(c.List))
}
