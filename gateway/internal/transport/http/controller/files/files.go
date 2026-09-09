package files

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
	filesservice "github.com/amirhossein2831/chat_bot/gateway/internal/service/files"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/dto"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/httpio"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/mapper"
	mw "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/middleware"
)

type Controller struct {
	service *filesservice.Service
}

func NewController(service *filesservice.Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) Presign(r *http.Request) (*httpio.Response, error) {
	var req dto.PresignRequest
	if err := httpio.Decode(r, &req); err != nil {
		return nil, err
	}

	upload, err := c.service.Presign(r.Context(), mw.UserFrom(r.Context()).ID,
		req.Filename, req.MIME, req.Size)
	if err != nil {
		return nil, err
	}

	return httpio.JSON(http.StatusCreated, dto.PresignResponse{
		AttachmentID: upload.AttachmentID.String(),
		URL:          upload.URL,
		ExpiresAt:    upload.ExpiresAt.Format(time.RFC3339),
	}), nil
}

func (c *Controller) Confirm(r *http.Request) (*httpio.Response, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}

	attachment, err := c.service.Confirm(r.Context(), id, mw.UserFrom(r.Context()).ID)
	if err != nil {
		return nil, err
	}

	return httpio.JSON(http.StatusOK, mapper.ToAttachmentResponse(attachment)), nil
}

func (c *Controller) Get(r *http.Request) (*httpio.Response, error) {
	id, err := pathID(r)
	if err != nil {
		return nil, err
	}

	attachment, err := c.service.Get(r.Context(), id, mw.UserFrom(r.Context()).ID)
	if err != nil {
		return nil, err
	}

	return httpio.JSON(http.StatusOK, mapper.ToAttachmentResponse(attachment)), nil
}

func pathID(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, constants.ErrBadRequest
	}
	return id, nil
}
