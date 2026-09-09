package httpio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/amirhossein2831/chat_bot/gateway/internal/constants"
)

const maxBodyBytes = 1 << 20

var validate = validator.New()

type Handler func(*http.Request) (*Response, error)

func Wrap(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := h(r)
		if err != nil {
			Error(w, err)
			return
		}
		if res == nil {
			Error(w, errors.New("handler returned no response"))
			return
		}
		res.write(w)
	}
}

func Decode(r *http.Request, target any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("%w: %s", constants.ErrBadRequest, err)
	}

	if err := validate.Struct(target); err != nil {
		return err
	}
	return nil
}

func Error(w http.ResponseWriter, err error) {
	status, payload := mapError(err)
	JSON(status, payload).write(w)
}
