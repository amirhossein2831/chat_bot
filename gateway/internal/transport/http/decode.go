package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
)

const maxBodyBytes = 1 << 20

var validate = validator.New()

func Decode(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("%w: %s", constants.ErrBadRequest, err)
	}

	if err := validate.Struct(target); err != nil {
		return fmt.Errorf("%w: %s", constants.ErrBadRequest, err)
	}
	return nil
}
