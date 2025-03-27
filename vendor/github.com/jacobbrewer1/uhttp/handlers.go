package uhttp

import (
	"errors"
	"net/http"
)

var (
	errNotFound         = errors.New("not found")
	errMethodNotAllowed = errors.New("method not allowed")
	errUnauthorized     = errors.New("unauthorized")
)

// WrapHandler wraps the handler with the specified middlewares, making the execution order the inverse of the parameter declaration.
func WrapHandler(handler http.HandlerFunc, middlewares ...MiddlewareFunc) http.Handler {
	var wrappedHandler http.Handler = handler
	for _, middleware := range middlewares {
		if middleware == nil {
			continue
		}
		wrappedHandler = middleware(wrappedHandler)
	}
	return wrappedHandler
}

// NotFoundHandler returns a handler that returns a 404 response.
func NotFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw, ok := w.(*ResponseWriter)
		if !ok {
			rw = NewResponseWriter(w,
				WithDefaultStatusCode(http.StatusNotFound),
				WithDefaultHeader(HeaderRequestID, RequestIDFromContext(GenerateOrCopyRequestID(r.Context(), r))),
				WithDefaultHeader(HeaderContentType, ContentTypeJSON),
			)
		}

		details := []any{
			"method: " + r.Method,
			"path: " + r.URL.Path,
		}

		if r.URL.RawQuery != "" {
			details = append(details, "query: "+r.URL.RawQuery)
		}

		msg := NewHTTPError(http.StatusNotFound, errNotFound, details...)

		// Is there a request ID in the context?
		reqId := RequestIDFromContext(r.Context())
		if reqId == "" {
			reqId = RequestIDFromContext(GenerateRequestIDToContext(r))
		}

		msg.RequestId = reqId
		rw.Header().Set(HeaderRequestID, reqId)
		MustEncode(rw, http.StatusNotFound, msg)
	}
}

// MethodNotAllowedHandler returns a handler that returns a 405 response.
func MethodNotAllowedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw, ok := w.(*ResponseWriter)
		if !ok {
			rw = NewResponseWriter(w,
				WithDefaultStatusCode(http.StatusMethodNotAllowed),
				WithDefaultHeader(HeaderRequestID, RequestIDFromContext(GenerateOrCopyRequestID(r.Context(), r))),
				WithDefaultHeader(HeaderContentType, ContentTypeJSON),
			)
		}

		details := []any{
			"method: " + r.Method,
			"path: " + r.URL.Path,
		}

		if r.URL.RawQuery != "" {
			details = append(details, "query: "+r.URL.RawQuery)
		}

		msg := NewHTTPError(http.StatusMethodNotAllowed, errMethodNotAllowed, details...)

		// Is there a request ID in the context?
		reqId := RequestIDFromContext(r.Context())
		if reqId == "" {
			reqId = RequestIDFromContext(GenerateRequestIDToContext(r))
		}

		msg.RequestId = reqId
		rw.Header().Set(HeaderRequestID, reqId)
		MustEncode(rw, http.StatusMethodNotAllowed, msg)
	}
}

// UnauthorizedHandler returns a handler that returns a 401 response.
func UnauthorizedHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw, ok := w.(*ResponseWriter)
		if !ok {
			rw = NewResponseWriter(w,
				WithDefaultStatusCode(http.StatusUnauthorized),
				WithDefaultHeader(HeaderRequestID, RequestIDFromContext(GenerateOrCopyRequestID(r.Context(), r))),
				WithDefaultHeader(HeaderContentType, ContentTypeJSON),
			)
		}

		details := []any{
			"method: " + r.Method,
			"path: " + r.URL.Path,
		}

		if r.URL.RawQuery != "" {
			details = append(details, "query: "+r.URL.RawQuery)
		}

		msg := NewHTTPError(http.StatusUnauthorized, errUnauthorized, details...)

		// Is there a request ID in the context?
		reqId := RequestIDFromContext(r.Context())
		if reqId == "" {
			reqId = RequestIDFromContext(GenerateRequestIDToContext(r))
		}

		msg.RequestId = reqId
		rw.Header().Set(HeaderRequestID, reqId)
		MustEncode(rw, http.StatusUnauthorized, msg)
	}
}

func GenericErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	rw, ok := w.(*ResponseWriter)
	if !ok {
		rw = NewResponseWriter(w,
			WithDefaultStatusCode(http.StatusBadRequest),
			WithDefaultHeader(HeaderRequestID, RequestIDFromContext(GenerateOrCopyRequestID(r.Context(), r))),
			WithDefaultHeader(HeaderContentType, ContentTypeJSON),
		)
	}

	msg := NewHTTPError(http.StatusBadRequest, err)

	// Is there a request ID in the context?
	reqId := RequestIDFromContext(r.Context())
	if reqId == "" {
		reqId = RequestIDFromContext(GenerateRequestIDToContext(r))
	}

	msg.RequestId = reqId
	rw.Header().Set(HeaderRequestID, reqId)
	MustEncode(rw, http.StatusBadRequest, msg)
}
