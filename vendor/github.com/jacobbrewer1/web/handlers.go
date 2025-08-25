package web

import (
	"net/http"

	"github.com/gorilla/mux"
)

// WrapHandler wraps an http.HandlerFunc with the provided middlewares.
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
