package router

import "github.com/go-chi/chi/v5"

// Routers are the scopes a controller can mount on. Root is outside the API
// prefix and unauthenticated, for probes and load balancers.
type Routers struct {
	Root      chi.Router
	Public    chi.Router
	Protected chi.Router
}

// Registrar lets a controller mount its own routes. Adding a controller means
// adding an implementation, never editing the router.
type Registrar interface {
	Routes(Routers)
}
