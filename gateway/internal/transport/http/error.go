package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
)

type errorBody struct {
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

// WriteError is the only place an error becomes a status code. Services never
// import net/http.
func WriteError(w http.ResponseWriter, err error) {
	for _, m := range mappings {
		if errors.Is(err, m.err) {
			WriteJSON(w, m.status, errorBody{Error: err.Error(), Code: m.code})
			return
		}
	}
	// An unmapped error carries DSNs, table names and upstream detail, so the
	// client is told nothing beyond the code.
	WriteJSON(w, http.StatusInternalServerError, errorBody{
		Error: "internal_error",
		Code:  "internal_error",
	})
}

func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
