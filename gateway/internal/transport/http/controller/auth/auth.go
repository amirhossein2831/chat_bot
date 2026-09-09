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

func (c *Controller) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := httpio.Decode(w, r, &req); err != nil {
		httpio.Error(w, err)
		return
	}

	user, err := c.service.Register(r.Context(), mapper.ToUserModel(req), req.Password, req.Invite)
	if err != nil {
		httpio.Error(w, err)
		return
	}

	httpio.JSON(w, http.StatusCreated, mapper.ToUserResponse(user))
}

func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := httpio.Decode(w, r, &req); err != nil {
		httpio.Error(w, err)
		return
	}

	token, user, err := c.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		httpio.Error(w, err)
		return
	}

	http.SetCookie(w, c.cookie(token, c.ttl))
	httpio.JSON(w, http.StatusOK, mapper.ToUserResponse(user))
}

func (c *Controller) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		if err := c.service.Logout(r.Context(), cookie.Value); err != nil {
			httpio.Error(w, err)
			return
		}
	}

	http.SetCookie(w, c.cookie("", -time.Second))
	httpio.NoContent(w)
}

func (c *Controller) Me(w http.ResponseWriter, r *http.Request) {
	httpio.JSON(w, http.StatusOK, mapper.ToUserResponse(mw.UserFrom(r.Context())))
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
