package web

import (
	"net/http"

	"github.com/gorilla/mux"
)

// WrapHandler wraps a http.HandlerFunc with the provided middlewares.
// The execution order of the middlewares is the inverse of the parameter declaration.
// This means that the first middleware in the list will be executed last.
func WrapHandler(handler http.HandlerFunc, middlewares ...mux.MiddlewareFunc) http.Handler {
	var wrappedHandler http.Handler = handler
	for _, middleware := range middlewares {
		if middleware == nil {
			continue
		}
		wrappedHandler = middleware(wrappedHandler)
	}
	return wrappedHandler
}
