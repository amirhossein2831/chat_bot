package controller

import (
	"net/http"
	"time"

	"github.com/amirhossein2831/chat_bot/gateway/internal/config"
	"github.com/amirhossein2831/chat_bot/gateway/internal/dto"
	"github.com/amirhossein2831/chat_bot/gateway/internal/mapper"
	"github.com/amirhossein2831/chat_bot/gateway/internal/service/auth"
	httptransport "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http"
	mw "github.com/amirhossein2831/chat_bot/gateway/internal/transport/http/middleware"
)

const sessionCookie = "session"

type Auth struct {
	service *auth.Service
	secure  bool
	ttl     time.Duration
}

func NewAuth(service *auth.Service, cfg *config.Config) *Auth {
	return &Auth{
		service: service,
		secure:  cfg.IsProduction(),
		ttl:     cfg.Auth.SessionTTL(),
	}
}

func (c *Auth) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := decode(w, r, &req); err != nil {
		httptransport.WriteError(w, err)
		return
	}

	user, err := c.service.Register(r.Context(), mapper.ToUserModel(req), req.Password, req.Invite)
	if err != nil {
		httptransport.WriteError(w, err)
		return
	}

	httptransport.WriteJSON(w, http.StatusCreated, mapper.ToUserResponse(user))
}

func (c *Auth) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := decode(w, r, &req); err != nil {
		httptransport.WriteError(w, err)
		return
	}

	token, user, err := c.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		httptransport.WriteError(w, err)
		return
	}

	http.SetCookie(w, c.cookie(token, c.ttl))
	httptransport.WriteJSON(w, http.StatusOK, mapper.ToUserResponse(user))
}

func (c *Auth) Logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookie); err == nil {
		if err := c.service.Logout(r.Context(), cookie.Value); err != nil {
			httptransport.WriteError(w, err)
			return
		}
	}

	http.SetCookie(w, c.cookie("", -time.Second))
	w.WriteHeader(http.StatusNoContent)
}

func (c *Auth) Me(w http.ResponseWriter, r *http.Request) {
	httptransport.WriteJSON(w, http.StatusOK, mapper.ToUserResponse(mw.UserFrom(r.Context())))
}

func (c *Auth) cookie(value string, ttl time.Duration) *http.Cookie {
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
