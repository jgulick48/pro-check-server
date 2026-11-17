package models

type Response struct {
	Sensors []Sensor `json:"sensors"`
}

type Sensor struct {
	Address           string             `json:"address"`
	Name              string             `json:"name,omitempty"`
	SensorType        string             `json:"sensorType"`
	SensorReadQuality float64            `json:"sensorReadQuality"`
	RSSI              int                `json:"rssi"`
	BatteryLevel      int                `json:"batteryLevel"`
	BatteryVoltage    float64            `json:"batteryVoltage"`
	TempCelsius       float64            `json:"tempCelsius"`
	TempFahrenheit    float64            `json:"tempFahrenheit"`
	TankLevelMM       float64            `json:"tankLevelMM"`
	TankLevelInches   float64            `json:"tankLevelInches"`
	TankLevelPercent  map[string]float64 `json:"tankLevelPercent"`
	TankType          string             `json:"tankType,omitempty"`
	LevelPercent      *float64           `json:"levelPercent,omitempty"`
	Gallons           *float64           `json:"gallons,omitempty"`
	LastUpdated       string             `json:"lastUpdated"`
	AgeSeconds        float64            `json:"ageSeconds"`
}
