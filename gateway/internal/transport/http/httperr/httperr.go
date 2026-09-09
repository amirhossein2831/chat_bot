package httperr

import (
	"errors"
	"net/http"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
)

const internalCode = "internal_error"

type Payload struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

type mapping struct {
	err    error
	status int
	code   string
}

var mappings = []mapping{
	{constants.ErrNotFound, http.StatusNotFound, "not_found"},
	{constants.ErrUnauthorized, http.StatusUnauthorized, "unauthorized"},
	{constants.ErrForbidden, http.StatusForbidden, "forbidden"},
	{constants.ErrConflict, http.StatusConflict, "conflict"},
	{constants.ErrBadRequest, http.StatusBadRequest, "bad_request"},
	{constants.ErrInvalidTransition, http.StatusConflict, "invalid_state"},
	{constants.ErrQuotaExceeded, http.StatusTooManyRequests, "quota_exceeded"},
	{constants.ErrModelUnavailable, http.StatusServiceUnavailable, "model_unavailable"},
	{constants.ErrContextTooLarge, http.StatusRequestEntityTooLarge, "context_too_large"},
}

func Map(err error) (int, Payload) {
	for _, m := range mappings {
		if errors.Is(err, m.err) {
			return m.status, Payload{Error: err.Error(), Code: m.code}
		}
	}
	return http.StatusInternalServerError, Payload{Error: internalCode, Code: internalCode}
}
