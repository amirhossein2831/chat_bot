package conversation

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
	conversationservice "github.com/amirhossein2831/chat_bot/gateway/internal/service/conversation"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/dto"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/httpio"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/mapper"
	mw "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/middleware"
)

type Controller struct {
	service *conversationservice.Service
}

func NewController(service *conversationservice.Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) Create(r *http.Request) (*httpio.Response, error) {
	var req dto.CreateConversationRequest
	if err := httpio.Decode(r, &req); err != nil {
		return nil, err
	}

	conversation, err := c.service.Create(r.Context(), mw.UserFrom(r.Context()).ID, req.Title, req.Model)
	if err != nil {
		return nil, err
	}

	return httpio.JSON(http.StatusCreated, mapper.ToConversationResponse(conversation)), nil
}

func (c *Controller) List(r *http.Request) (*httpio.Response, error) {
	conversations, err := c.service.List(r.Context(),
		mw.UserFrom(r.Context()).ID, r.URL.Query().Get("q"))
	if err != nil {
		return nil, err
	}

	return httpio.JSON(http.StatusOK, mapper.ToConversationResponses(conversations)), nil
}

func (c *Controller) Messages(r *http.Request) (*httpio.Response, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}

	thread, err := c.service.Thread(r.Context(), id, mw.UserFrom(r.Context()).ID)
	if err != nil {
		return nil, err
	}

	return httpio.JSON(http.StatusOK, mapper.ToThreadResponse(thread)), nil
}

func (c *Controller) Rename(r *http.Request) (*httpio.Response, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}

	var req dto.RenameConversationRequest
	if err := httpio.Decode(r, &req); err != nil {
		return nil, err
	}

	if err := c.service.Rename(r.Context(), id, mw.UserFrom(r.Context()).ID, req.Title); err != nil {
		return nil, err
	}

	return httpio.NoContent(), nil
}

func (c *Controller) Archive(r *http.Request) (*httpio.Response, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}

	if err := c.service.Archive(r.Context(), id, mw.UserFrom(r.Context()).ID); err != nil {
		return nil, err
	}

	return httpio.NoContent(), nil
}

func (c *Controller) SwitchBranch(r *http.Request) (*httpio.Response, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}

	var req dto.SwitchBranchRequest
	if err := httpio.Decode(r, &req); err != nil {
		return nil, err
	}

	messageID, err := uuid.Parse(req.MessageID)
	if err != nil {
		return nil, constants.ErrBadRequest
	}

	if err := c.service.SwitchBranch(r.Context(), id, mw.UserFrom(r.Context()).ID, messageID); err != nil {
		return nil, err
	}

	return httpio.NoContent(), nil
}

func pathID(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, constants.ErrBadRequest
	}
	return id, nil
}
