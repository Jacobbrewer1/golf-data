package metrics

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// InstrumentDuration instruments a handler with the request duration, updating the supplied histogram.
func InstrumentDuration(metric *prometheus.HistogramVec, options ...promhttp.Option) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return promhttp.InstrumentHandlerDuration(metric, next, options...)
	}
}

// InstrumentCounter instruments a handler with the request counter, updating the supplied counter.
func InstrumentCounter(metric *prometheus.CounterVec, options ...promhttp.Option) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return promhttp.InstrumentHandlerCounter(metric, next, options...)
	}
}

// InstrumentRequestSize instruments a handler with the request size, updating the supplied histogram.
func InstrumentRequestSize(metric *prometheus.HistogramVec, options ...promhttp.Option) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return promhttp.InstrumentHandlerRequestSize(metric, next, options...)
	}
}

// InstrumentResponseSize instruments a handler with the response size, updating the supplied histogram.
func InstrumentResponseSize(metric *prometheus.HistogramVec, options ...promhttp.Option) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return promhttp.InstrumentHandlerResponseSize(metric, next, options...)
	}
}
