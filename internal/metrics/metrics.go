// Package metrics exposes sensor readings as Prometheus gauges.
package metrics

import (
	"github.com/jgulick48/mopeka_pro_check"
	"github.com/jgulick48/pro-check-server/internal/config"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	labels = []string{"address", "name", "tank_type"}

	levelMM      = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mopeka_level_mm", Help: "Liquid level above the sensor in mm."}, labels)
	levelPercent = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mopeka_level_percent", Help: "Fill by volume for the configured tank type."}, labels)
	gallons      = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mopeka_level_gallons", Help: "Liquid volume for tank types with a known capacity."}, labels)
	tempC        = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mopeka_temperature_celsius", Help: "Sensor temperature."}, labels)
	batteryV     = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mopeka_battery_volts", Help: "Sensor battery voltage."}, labels)
	quality      = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mopeka_read_quality", Help: "Reading quality, 0 (worst) to 3 (best)."}, labels)
	rssi         = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mopeka_rssi_dbm", Help: "Signal strength of the last advertisement."}, labels)
	lastSeen     = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mopeka_last_seen_timestamp_seconds", Help: "Unix time of the last advertisement."}, labels)
	adverts      = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mopeka_scanner_advertisements_total", Help: "Advertisements with manufacturer data seen by the adapter (any device)."})
	running      = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mopeka_scanner_running", Help: "1 while the BLE scan loop runs."})
)

func init() {
	prometheus.MustRegister(levelMM, levelPercent, gallons, tempC, batteryV, quality, rssi, lastSeen, adverts, running)
}

// Update refreshes all gauges from the scanner.
func Update(s *mopeka_pro_check.Scanner, cfg config.Config) {
	adverts.Set(float64(s.AdvertisementsSeen()))
	if s.Running() {
		running.Set(1)
	} else {
		running.Set(0)
	}
	for _, d := range s.GetDevices() {
		tank := cfg.TankFor(d.GetAddress())
		l := prometheus.Labels{"address": d.GetAddress(), "name": tank.Name, "tank_type": tank.Type}
		levelMM.With(l).Set(d.GetTankLevelMM())
		tempC.With(l).Set(d.GetTempCelsius())
		batteryV.With(l).Set(d.GetBatteryVoltage())
		quality.With(l).Set(d.GetReadQuality())
		rssi.With(l).Set(float64(d.GetRSSI()))
		lastSeen.With(l).Set(float64(d.GetReadingTime().Unix()))
		if tank.Type != "" {
			p := d.GetLevelPercent(tank.Type)
			levelPercent.With(l).Set(p)
			if gal, ok := mopeka_pro_check.TankVolumeGallons[tank.Type]; ok {
				gallons.With(l).Set(gal * p / 100)
			}
		}
	}
}
