package models

import (
	"time"

	"github.com/jgulick48/mopeka_pro_check"
)

// FromDevice builds the API/MQTT view of one reading. tankType and name come
// from the configuration and may be empty.
func FromDevice(d mopeka_pro_check.MopekaProCheck, tankType, name string) Sensor {
	s := Sensor{
		Address:           d.GetAddress(),
		Name:              name,
		SensorType:        d.GetSensorType(),
		SensorReadQuality: d.GetReadQuality(),
		RSSI:              d.GetRSSI(),
		BatteryLevel:      d.GetBatteryLevel(),
		BatteryVoltage:    d.GetBatteryVoltage(),
		TempCelsius:       d.GetTempCelsius(),
		TempFahrenheit:    d.GetTempFahrenheit(),
		TankLevelMM:       d.GetTankLevelMM(),
		TankLevelInches:   d.GetTankLevelInches(),
		TankLevelPercent:  map[string]float64{},
		TankType:          tankType,
		LastUpdated:       d.GetReadingTime().Format(time.RFC3339),
		AgeSeconds:        time.Since(d.GetReadingTime()).Seconds(),
	}
	for t := range mopeka_pro_check.TankTypes {
		s.TankLevelPercent[t] = d.GetLevelPercent(t)
	}
	if tankType != "" {
		p := d.GetLevelPercent(tankType)
		s.LevelPercent = &p
		if gal, ok := mopeka_pro_check.TankVolumeGallons[tankType]; ok {
			g := gal * p / 100
			s.Gallons = &g
		}
	}
	return s
}
