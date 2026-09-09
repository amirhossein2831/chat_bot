package httpio

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
)

const internalCode = "internal_error"

type errorPayload struct {
	Error  string       `json:"error"`
	Code   string       `json:"code"`
	Fields []fieldError `json:"fields,omitempty"`
}

type fieldError struct {
	Field string `json:"field"`
	Rule  string `json:"rule"`
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
	{constants.ErrMethodNotAllowed, http.StatusMethodNotAllowed, "method_not_allowed"},
	{constants.ErrInvalidTransition, http.StatusConflict, "invalid_state"},
	{constants.ErrQuotaExceeded, http.StatusTooManyRequests, "quota_exceeded"},
	{constants.ErrModelUnavailable, http.StatusServiceUnavailable, "model_unavailable"},
	{constants.ErrContextTooLarge, http.StatusRequestEntityTooLarge, "context_too_large"},
}

func mapError(err error) (int, errorPayload) {
	var invalid validator.ValidationErrors
	if errors.As(err, &invalid) {
		return http.StatusBadRequest, errorPayload{
			Error:  "validation failed",
			Code:   "validation_failed",
			Fields: fieldErrors(invalid),
		}
	}

	for _, m := range mappings {
		if errors.Is(err, m.err) {
			return m.status, errorPayload{Error: err.Error(), Code: m.code}
		}
	}

	// An unmapped error carries DSNs, table names and upstream detail, so its
	// message is replaced rather than forwarded.
	return http.StatusInternalServerError, errorPayload{Error: internalCode, Code: internalCode}
}

// fieldErrors names the JSON field, not the Go struct field, so the client
// never learns the internal type names.
func fieldErrors(invalid validator.ValidationErrors) []fieldError {
	out := make([]fieldError, 0, len(invalid))
	for _, e := range invalid {
		out = append(out, fieldError{
			Field: strings.ToLower(e.Field()),
			Rule:  e.Tag(),
		})
	}
	return out
}
