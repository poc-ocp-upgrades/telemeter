package http

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func DebugRoutes(mux *http.ServeMux) *http.ServeMux {
	_logClusterCodePath()
	defer _logClusterCodePath()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.Handle("/debug/pprof/block", pprof.Handler("block"))
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	return mux
}
func HealthRoutes(mux *http.ServeMux) *http.ServeMux {
	_logClusterCodePath()
	defer _logClusterCodePath()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	mux.HandleFunc("/healthz/ready", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	return mux
}
func MetricRoutes(mux *http.ServeMux) *http.ServeMux {
	_logClusterCodePath()
	defer _logClusterCodePath()
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}
func ReloadRoutes(mux *http.ServeMux, reload func() error) *http.ServeMux {
	_logClusterCodePath()
	defer _logClusterCodePath()
	mux.HandleFunc("/-/reload", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if err := reload(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	return mux
}
