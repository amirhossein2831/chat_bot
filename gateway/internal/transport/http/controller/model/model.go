package model

import (
	"net/http"

	"github.com/amirhossein2831/chat_bot/gateway/internal/client/registry"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/httpio"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/mapper"
)

type Controller struct {
	registry *registry.Registry
}

func NewController(registry *registry.Registry) *Controller {
	return &Controller{registry: registry}
}

func (c *Controller) List(*http.Request) (*httpio.Response, error) {
	return httpio.JSON(http.StatusOK, mapper.ToModelResponses(c.registry.Models())), nil
}
