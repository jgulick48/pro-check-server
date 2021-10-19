package models

type Response struct {
	Sensors []Sensor `json:"sensors"`
}

type Sensor struct {
	Address          string             `json:"address"`
	SensorType       string             `json:"sensorType"`
	BatteryLevel     int                `json:"batteryLevel"`
	BatteryVoltage   float64            `json:"batteryVoltage"`
	TempCelsius      float64            `json:"tempCelsius"`
	TempFahrenheit   float64            `json:"tempFahrenheit"`
	TankLevelMM      float64            `json:"tankLevelMM"`
	TankLevelInches  float64            `json:"tankLevelInches"`
	TankLevelPercent map[string]float64 `json:"tankLevelPercent"`
	LastUpdated      string             `json:"lastUpdated"`
}
