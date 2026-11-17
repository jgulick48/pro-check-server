package api

import (
	"encoding/json"
	"net/http"

	"github.com/jgulick48/mopeka_pro_check"
	"github.com/jgulick48/pro-check-server/internal/config"
	"github.com/jgulick48/pro-check-server/internal/models"
)

type GetAllSensorsHandler struct {
	tankSensors *mopeka_pro_check.Scanner
	cfg         config.Config
}

func (h *GetAllSensorsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sensors := h.tankSensors.GetDevices()
	response := models.Response{Sensors: make([]models.Sensor, 0, len(sensors))}
	for _, sensor := range sensors {
		tank := h.cfg.TankFor(sensor.GetAddress())
		response.Sensors = append(response.Sensors, models.FromDevice(sensor, tank.Type, tank.Name))
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func NewGetAllSensorsHandler(tankSensors *mopeka_pro_check.Scanner, cfg config.Config) http.Handler {
	return &GetAllSensorsHandler{tankSensors: tankSensors, cfg: cfg}
}
