package uhttp

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MiddlewareFunc is a middleware function.
type MiddlewareFunc = mux.MiddlewareFunc

// InstrumentDuration instruments a handler with the request duration, updating the supplied histogram.
func InstrumentDuration(metric *prometheus.HistogramVec) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return promhttp.InstrumentHandlerDuration(metric, next)
	}
}

// InstrumentCounter instruments a handler with the request counter, updating the supplied counter.
func InstrumentCounter(metric *prometheus.CounterVec) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return promhttp.InstrumentHandlerCounter(metric, next)
	}
}

// InstrumentRequestSize instruments a handler with the request size, updating the supplied histogram.
func InstrumentRequestSize(metric *prometheus.HistogramVec) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return promhttp.InstrumentHandlerRequestSize(metric, next)
	}
}

// InstrumentResponseSize instruments a handler with the response size, updating the supplied histogram.
func InstrumentResponseSize(metric *prometheus.HistogramVec) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return promhttp.InstrumentHandlerResponseSize(metric, next)
	}
}
