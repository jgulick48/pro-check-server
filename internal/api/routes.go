package api

import (
	"github.com/gorilla/mux"
	"github.com/jgulick48/mopeka_pro_check"
)

func NewAPIRouter(tankSensors mopeka_pro_check.Scanner) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/sensors", NewGetAllSensorsHandler(tankSensors).ServeHTTP)
	return r
}
