package health

import (
	"net/http"

	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/httpio"
)

type Controller struct{}

func NewController() *Controller { return &Controller{} }

func (c *Controller) Live(*http.Request) (*httpio.Response, error) {
	return httpio.JSON(http.StatusOK, map[string]string{"status": "ok"}), nil
}
