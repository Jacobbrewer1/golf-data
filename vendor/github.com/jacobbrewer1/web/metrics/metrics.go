package metrics

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// InstrumentDuration instruments an HTTP handler with request duration metrics.
//
// This function wraps the provided HTTP handler to measure the duration of each request
// and updates the supplied Prometheus histogram with the recorded durations.
func InstrumentDuration(metric *prometheus.HistogramVec, options ...promhttp.Option) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return promhttp.InstrumentHandlerDuration(metric, next, options...)
	}
}

// InstrumentCounter instruments an HTTP handler with request count metrics.
//
// This function wraps the provided HTTP handler to count the number of requests
// and updates the supplied Prometheus counter with the recorded counts.
func InstrumentCounter(metric *prometheus.CounterVec, options ...promhttp.Option) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return promhttp.InstrumentHandlerCounter(metric, next, options...)
	}
}

// InstrumentRequestSize instruments an HTTP handler with request size metrics.
//
// This function wraps the provided HTTP handler to measure the size of incoming requests
// and updates the supplied Prometheus histogram with the recorded sizes.
func InstrumentRequestSize(metric *prometheus.HistogramVec, options ...promhttp.Option) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return promhttp.InstrumentHandlerRequestSize(metric, next, options...)
	}
}

// InstrumentResponseSize instruments an HTTP handler with response size metrics.
//
// This function wraps the provided HTTP handler to measure the size of outgoing responses
// and updates the supplied Prometheus histogram with the recorded sizes.
func InstrumentResponseSize(metric *prometheus.HistogramVec, options ...promhttp.Option) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return promhttp.InstrumentHandlerResponseSize(metric, next, options...)
	}
}
