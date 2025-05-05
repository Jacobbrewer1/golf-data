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
//
// Parameters:
//   - metric (*prometheus.HistogramVec): A Prometheus histogram vector to record request durations.
//   - options (...promhttp.Option): Additional options for customizing the instrumentation.
//
// Returns:
//   - mux.MiddlewareFunc: A middleware function that instruments the handler with request duration metrics.
func InstrumentDuration(metric *prometheus.HistogramVec, options ...promhttp.Option) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return promhttp.InstrumentHandlerDuration(metric, next, options...)
	}
}

// InstrumentCounter instruments an HTTP handler with request count metrics.
//
// This function wraps the provided HTTP handler to count the number of requests
// and updates the supplied Prometheus counter with the recorded counts.
//
// Parameters:
//   - metric (*prometheus.CounterVec): A Prometheus counter vector to record request counts.
//   - options (...promhttp.Option): Additional options for customizing the instrumentation.
//
// Returns:
//   - mux.MiddlewareFunc: A middleware function that instruments the handler with request count metrics.
func InstrumentCounter(metric *prometheus.CounterVec, options ...promhttp.Option) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return promhttp.InstrumentHandlerCounter(metric, next, options...)
	}
}

// InstrumentRequestSize instruments an HTTP handler with request size metrics.
//
// This function wraps the provided HTTP handler to measure the size of incoming requests
// and updates the supplied Prometheus histogram with the recorded sizes.
//
// Parameters:
//   - metric (*prometheus.HistogramVec): A Prometheus histogram vector to record request sizes.
//   - options (...promhttp.Option): Additional options for customizing the instrumentation.
//
// Returns:
//   - mux.MiddlewareFunc: A middleware function that instruments the handler with request size metrics.
func InstrumentRequestSize(metric *prometheus.HistogramVec, options ...promhttp.Option) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return promhttp.InstrumentHandlerRequestSize(metric, next, options...)
	}
}

// InstrumentResponseSize instruments an HTTP handler with response size metrics.
//
// This function wraps the provided HTTP handler to measure the size of outgoing responses
// and updates the supplied Prometheus histogram with the recorded sizes.
//
// Parameters:
//   - metric (*prometheus.HistogramVec): A Prometheus histogram vector to record response sizes.
//   - options (...promhttp.Option): Additional options for customizing the instrumentation.
//
// Returns:
//   - mux.MiddlewareFunc: A middleware function that instruments the handler with response size metrics.
func InstrumentResponseSize(metric *prometheus.HistogramVec, options ...promhttp.Option) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return promhttp.InstrumentHandlerResponseSize(metric, next, options...)
	}
}
