package web

import (
	"net/http"

	"github.com/gorilla/mux"
)

// WrapHandler wraps an http.HandlerFunc with the provided middlewares.
//
// This function takes an http.HandlerFunc and a variadic list of mux.MiddlewareFunc.
// It applies the middlewares in reverse order of their declaration, meaning the first
// middleware in the list will be executed last.
//
// Parameters:
//   - handler: The http.HandlerFunc to be wrapped.
//   - middlewares: A variadic list of mux.MiddlewareFunc to be applied to the handler.
//
// Returns:
//   - http.Handler: The resulting http.Handler with all middlewares applied.
//
// Behavior:
//   - If a middleware in the list is nil, it is skipped.
//   - The middlewares are applied in reverse order, ensuring the first middleware
//     in the list wraps the entire chain and is executed last.
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
