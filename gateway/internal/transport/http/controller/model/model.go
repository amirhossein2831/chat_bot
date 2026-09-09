package model

import (
	"net/http"

	"github.com/amirhossein2831/chat_bot/gateway/internal/service/catalogue"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/httpio"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/mapper"
	mw "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/middleware"
)

type Controller struct {
	service *catalogue.Service
}

func NewController(service *catalogue.Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) List(r *http.Request) (*httpio.Response, error) {
	models := c.service.Available(r.Context(), mw.UserFrom(r.Context()))
	return httpio.JSON(http.StatusOK, mapper.ToModelResponses(models)), nil
}
