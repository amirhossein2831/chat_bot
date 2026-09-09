package auth

import (
	"net/http"
	"time"

	"github.com/amirhossein2831/chat_bot/gateway/internal/config"
	authservice "github.com/amirhossein2831/chat_bot/gateway/internal/service/auth"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/dto"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/httpio"
	"github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/mapper"
	mw "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/middleware"
)

const sessionCookie = "session"

type Controller struct {
	service *authservice.Service
	secure  bool
	ttl     time.Duration
}

func NewController(service *authservice.Service, cfg *config.Config) *Controller {
	return &Controller{
		service: service,
		secure:  cfg.IsProduction(),
		ttl:     cfg.Auth.SessionTTL(),
	}
}

func (c *Controller) Register(r *http.Request) (*httpio.Response, error) {
	var req dto.RegisterRequest
	if err := httpio.Decode(r, &req); err != nil {
		return nil, err
	}

	user, err := c.service.Register(r.Context(), mapper.ToUserModel(req), req.Password, req.Invite)
	if err != nil {
		return nil, err
	}

	return httpio.JSON(http.StatusCreated, mapper.ToUserResponse(user)), nil
}

func (c *Controller) Login(r *http.Request) (*httpio.Response, error) {
	var req dto.LoginRequest
	if err := httpio.Decode(r, &req); err != nil {
		return nil, err
	}

	token, user, err := c.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return httpio.JSON(http.StatusOK, mapper.ToUserResponse(user)).
		WithCookie(c.cookie(token, c.ttl)), nil
}

func (c *Controller) Logout(r *http.Request) (*httpio.Response, error) {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		if err := c.service.Logout(r.Context(), cookie.Value); err != nil {
			return nil, err
		}
	}

	return httpio.NoContent().WithCookie(c.cookie("", -time.Second)), nil
}

func (c *Controller) Me(r *http.Request) (*httpio.Response, error) {
	return httpio.JSON(http.StatusOK, mapper.ToUserResponse(mw.UserFrom(r.Context()))), nil
}

func (c *Controller) cookie(value string, ttl time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:     sessionCookie,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   c.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(ttl.Seconds()),
	}
}
