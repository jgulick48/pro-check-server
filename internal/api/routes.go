package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jgulick48/mopeka_pro_check"
	"github.com/jgulick48/pro-check-server/internal/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewAPIRouter(tankSensors *mopeka_pro_check.Scanner, cfg config.Config) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/sensors", NewGetAllSensorsHandler(tankSensors, cfg).ServeHTTP)
	r.Handle("/metrics", promhttp.Handler())
	r.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		if !tankSensors.Running() {
			http.Error(w, "scanner stopped", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok\n"))
	})
	return r
}
