package memory

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
	memoryservice "github.com/amirhossein2831/chat_bot/gateway/internal/service/memory"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/dto"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/httpio"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/mapper"
	mw "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/middleware"
)

type Controller struct {
	service *memoryservice.Service
}

func NewController(service *memoryservice.Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) Settings(r *http.Request) (*httpio.Response, error) {
	instructions, err := c.service.Instructions(r.Context(), mw.UserFrom(r.Context()).ID)
	if err != nil {
		return nil, err
	}
	return httpio.JSON(http.StatusOK, dto.SettingsResponse{Instructions: instructions}), nil
}

func (c *Controller) SetSettings(r *http.Request) (*httpio.Response, error) {
	var req dto.InstructionsRequest
	if err := httpio.Decode(r, &req); err != nil {
		return nil, err
	}
	if err := c.service.SetInstructions(r.Context(), mw.UserFrom(r.Context()).ID, req.Instructions); err != nil {
		return nil, err
	}
	return httpio.NoContent(), nil
}

func (c *Controller) List(r *http.Request) (*httpio.Response, error) {
	facts, err := c.service.Facts(r.Context(), mw.UserFrom(r.Context()).ID)
	if err != nil {
		return nil, err
	}
	return httpio.JSON(http.StatusOK, mapper.ToFactResponses(facts)), nil
}

func (c *Controller) Add(r *http.Request) (*httpio.Response, error) {
	var req dto.AddFactRequest
	if err := httpio.Decode(r, &req); err != nil {
		return nil, err
	}
	fact, err := c.service.Add(r.Context(), mw.UserFrom(r.Context()).ID, req.Text)
	if err != nil {
		return nil, err
	}
	return httpio.JSON(http.StatusCreated, mapper.ToFactResponse(fact)), nil
}

func (c *Controller) Delete(r *http.Request) (*httpio.Response, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return nil, constants.ErrBadRequest
	}
	if err := c.service.Delete(r.Context(), id, mw.UserFrom(r.Context()).ID); err != nil {
		return nil, err
	}
	return httpio.NoContent(), nil
}
