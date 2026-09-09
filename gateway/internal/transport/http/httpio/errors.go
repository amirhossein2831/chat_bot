package httpio

import (
	"errors"
	"net/http"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
)

type errorPayload struct {
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

const internalCode = "internal_error"

// mapError decides what the client is allowed to see. An unmapped error
// carries DSNs, table names and upstream detail, so its message is replaced
// rather than forwarded.
func mapError(err error) (int, errorPayload) {
	for _, m := range mappings {
		if errors.Is(err, m.err) {
			return m.status, errorPayload{Error: err.Error(), Code: m.code}
		}
	}
	return http.StatusInternalServerError, errorPayload{Error: internalCode, Code: internalCode}
}
