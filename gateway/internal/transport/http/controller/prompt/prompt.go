package prompt

import (
	"net/http"

	promptservice "github.com/amirhossein2831/chat_bot/gateway/internal/service/prompts"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/httpio"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/mapper"
	mw "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/middleware"
)

type Controller struct {
	service *promptservice.Service
}

func NewController(service *promptservice.Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) List(r *http.Request) (*httpio.Response, error) {
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = mw.UserFrom(r.Context()).Locale
	}

	prompts, err := c.service.List(r.Context(), locale, r.URL.Query().Get("category"))
	if err != nil {
		return nil, err
	}

	return httpio.JSON(http.StatusOK, mapper.ToPromptResponses(prompts)), nil
}
