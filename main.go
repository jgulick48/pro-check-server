package main

import (
	"github.com/jgulick48/pro-check-server/internal/api"
	"log"
	"net/http"
	"time"

	"github.com/jgulick48/mopeka_pro_check"
)

func main() {
	tankSensors := mopeka_pro_check.NewScanner(20 * time.Second)
	tankSensors.StartScan()
	go func() {
		t := time.NewTicker(time.Minute)
		for range t.C {
			if tankSensors.GetLastUpdateTime().Before(time.Now().Add(-5 * time.Minute)) {
				log.Printf("No sensor data has been updated in the last 5 minutes. Restarting.")
				tankSensors.StopScan()
				tankSensors.StartScan()
			}
			if tankSensors.GetLastUpdateTime().Before(time.Now().Add(-10 * time.Minute)) {
				log.Fatalf("No sensor data has been updated in the last 10 minutes. Dieing.")
			}
		}
	}()
	time.Sleep(20 * time.Second)
	r := api.NewAPIRouter(tankSensors)
	srv := &http.Server{
		Handler: r,
		Addr:    "0.0.0.0:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
