package http

import (
	"net/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestDuration	= prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "http_request_duration_seconds", Help: "Tracks the latencies for HTTP requests."}, []string{"code", "handler", "method"})
	requestSize	= prometheus.NewSummaryVec(prometheus.SummaryOpts{Name: "http_request_size_bytes", Help: "Tracks the size of HTTP requests."}, []string{"code", "handler", "method"})
	requestsTotal	= prometheus.NewCounterVec(prometheus.CounterOpts{Name: "http_requests_total", Help: "Tracks the number of HTTP requests."}, []string{"code", "handler", "method"})
)

func init() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	prometheus.MustRegister(requestDuration, requestSize, requestsTotal)
}
func NewInstrumentedHandler(handlerName string, next http.Handler) http.Handler {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return promhttp.InstrumentHandlerDuration(requestDuration.MustCurryWith(prometheus.Labels{"handler": handlerName}), promhttp.InstrumentHandlerRequestSize(requestSize.MustCurryWith(prometheus.Labels{"handler": handlerName}), promhttp.InstrumentHandlerCounter(requestsTotal.MustCurryWith(prometheus.Labels{"handler": handlerName}), next)))
}
