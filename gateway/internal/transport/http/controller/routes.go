package controller

import "github.com/go-chi/chi/v5"

func (c *Auth) Routes(public, protected chi.Router) {
	public.Post("/auth/register", c.Register)
	public.Post("/auth/login", c.Login)

	protected.Post("/auth/logout", c.Logout)
	protected.Get("/me", c.Me)
}
