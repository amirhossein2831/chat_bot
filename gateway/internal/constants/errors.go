package constants

import "errors"

var (
	ErrNotFound           = errors.New("not found")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrConflict           = errors.New("conflict")
	ErrBadRequest         = errors.New("bad request")
	ErrMethodNotAllowed   = errors.New("method not allowed")
	ErrQuotaExceeded      = errors.New("quota exceeded")
	ErrModelUnavailable   = errors.New("model unavailable")
	ErrStorageUnavailable = errors.New("storage unavailable")
	ErrInvalidTransition  = errors.New("invalid state transition")
	ErrContextTooLarge    = errors.New("context too large")
)
