package router

import "github.com/go-chi/chi/v5"

// Registrar lets a controller mount its own routes. Adding a controller means
// adding an implementation, never editing the router.
type Registrar interface {
	Routes(public, protected chi.Router)
}
