package chat

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
	chatservice "github.com/amirhossein2831/chat_bot/gateway/internal/service/chat"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/dto"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/httpio"
	mw "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/middleware"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/sse"
)

type Controller struct {
	service *chatservice.Service
}

func NewController(service *chatservice.Service) *Controller {
	return &Controller{service: service}
}

// Send is a raw handler rather than an httpio.Handler: it writes an open
// stream, not a single response.
func (c *Controller) Send(w http.ResponseWriter, r *http.Request) {
	var req dto.SendMessageRequest
	if err := httpio.Decode(r, &req); err != nil {
		httpio.Error(w, err)
		return
	}

	conversationID, err := uuid.Parse(req.ConversationID)
	if err != nil {
		httpio.Error(w, constants.ErrBadRequest)
		return
	}

	var editOf *uuid.UUID
	if req.EditOf != "" {
		parsed, err := uuid.Parse(req.EditOf)
		if err != nil {
			httpio.Error(w, constants.ErrBadRequest)
			return
		}
		editOf = &parsed
	}

	attachments := make([]uuid.UUID, 0, len(req.AttachmentIDs))
	for _, id := range req.AttachmentIDs {
		parsed, err := uuid.Parse(id)
		if err != nil {
			httpio.Error(w, constants.ErrBadRequest)
			return
		}
		attachments = append(attachments, parsed)
	}

	started, err := c.service.Send(r.Context(), chatservice.SendCommand{
		UserID:         mw.UserFrom(r.Context()).ID,
		ConversationID: conversationID,
		Content:        req.Content,
		Model:          req.Model,
		EditOf:         editOf,
		AttachmentIDs:  attachments,
	})
	if err != nil {
		httpio.Error(w, err)
		return
	}

	rc := sse.Prepare(w)
	sse.Meta(w, rc, started.MessageID.String(), started.Truncated)
	sse.Pipe(w, r, rc, started.Deltas)
}

func (c *Controller) Regenerate(w http.ResponseWriter, r *http.Request) {
	messageID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpio.Error(w, constants.ErrBadRequest)
		return
	}

	var req dto.RegenerateRequest
	if err := httpio.Decode(r, &req); err != nil {
		httpio.Error(w, err)
		return
	}

	started, err := c.service.Regenerate(r.Context(), chatservice.RegenerateCommand{
		UserID:    mw.UserFrom(r.Context()).ID,
		MessageID: messageID,
		Model:     req.Model,
	})
	if err != nil {
		httpio.Error(w, err)
		return
	}

	rc := sse.Prepare(w)
	sse.Meta(w, rc, started.MessageID.String(), started.Truncated)
	sse.Pipe(w, r, rc, started.Deltas)
}

func (c *Controller) Reattach(w http.ResponseWriter, r *http.Request) {
	messageID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpio.Error(w, constants.ErrBadRequest)
		return
	}

	deltas, err := c.service.Reattach(messageID)
	if err != nil {
		httpio.Error(w, err)
		return
	}

	rc := sse.Prepare(w)
	sse.Meta(w, rc, messageID.String(), false)
	sse.Pipe(w, r, rc, deltas)
}

func (c *Controller) Stop(r *http.Request) (*httpio.Response, error) {
	messageID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return nil, constants.ErrBadRequest
	}
	if err := c.service.Stop(r.Context(), messageID); err != nil {
		return nil, err
	}
	return httpio.NoContent(), nil
}
