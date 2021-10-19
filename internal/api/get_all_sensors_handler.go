package api

import (
	"encoding/json"
	"github.com/jgulick48/mopeka_pro_check"
	"github.com/jgulick48/pro-check-server/internal/models"
	"net/http"
)

type GetAllSensorsHandler struct {
	tankSensors mopeka_pro_check.Scanner
}

func (h *GetAllSensorsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sensors := h.tankSensors.GetDevices()
	response := models.Response{Sensors: make([]models.Sensor, 0, len(sensors))}
	for _, sensor := range sensors {
		sensorData := models.Sensor{
			Address:         sensor.GetAddress(),
			SensorType:      sensor.GetSensorType(),
			BatteryLevel:    sensor.GetBatteryLevel(),
			BatteryVoltage:  sensor.GetBatteryVoltage(),
			TempCelsius:     sensor.GetTempCelsius(),
			TempFahrenheit:  sensor.GetTempFahrenheit(),
			TankLevelMM:     sensor.GetTankLevelMM(),
			TankLevelInches: sensor.GetTankLevelInches(),
			TankLevelPercent: map[string]float64{
				"20lb_v":  sensor.GetLevelPercent("20lb_v"),
				"30lb_v":  sensor.GetLevelPercent("30lb_v"),
				"40lb_v":  sensor.GetLevelPercent("40lb_v"),
				"100lb_v": sensor.GetLevelPercent("100lb_v"),
			},
			LastUpdated: sensor.GetReadingTime().String(),
		}
		response.Sensors = append(response.Sensors, sensorData)
	}

	json.NewEncoder(w).Encode(response)
}

func NewGetAllSensorsHandler(tankSensors mopeka_pro_check.Scanner) http.Handler {
	return &GetAllSensorsHandler{tankSensors: tankSensors}
}
