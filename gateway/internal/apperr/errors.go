package apperr

import "errors"

var (
	NotFound          = errors.New("not found")
	Unauthorized      = errors.New("unauthorized")
	Forbidden         = errors.New("forbidden")
	Conflict          = errors.New("conflict")
	BadRequest        = errors.New("bad request")
	QuotaExceeded     = errors.New("quota exceeded")
	ModelUnavailable  = errors.New("model unavailable")
	InvalidTransition = errors.New("invalid state transition")
	ContextTooLarge   = errors.New("context too large")
)
