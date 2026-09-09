package router

import "github.com/go-chi/chi/v5"

type Routers struct {
	Root      chi.Router
	Public    chi.Router
	Protected chi.Router
}

type Registrar interface {
	Routes(Routers)
}
