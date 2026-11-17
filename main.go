package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jgulick48/mopeka_pro_check"
	"github.com/jgulick48/pro-check-server/internal/api"
	"github.com/jgulick48/pro-check-server/internal/config"
	"github.com/jgulick48/pro-check-server/internal/metrics"
	"github.com/jgulick48/pro-check-server/internal/mqttpub"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	tankSensors := mopeka_pro_check.NewScannerForDevice(cfg.HCIDevice)
	tankSensors.StartScan()

	pub, err := mqttpub.New(cfg)
	if err != nil {
		log.Fatalf("mqtt: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Every 10 s: refresh metrics and publish new readings. Restart the scan if
	// the adapter reports no advertisements at all for 5 minutes (any device,
	// so a quiet or distant tank sensor doesn't trigger it); exit if the scan
	// loop has stopped, so the container restarts.
	go func() {
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		lastSeen, lastChange, restarts := tankSensors.AdvertisementsSeen(), time.Now(), 0
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
			if !tankSensors.Running() {
				log.Fatalf("BLE scanner stopped: %v", tankSensors.Err())
			}
			if n := tankSensors.AdvertisementsSeen(); n != lastSeen {
				lastSeen, lastChange, restarts = n, time.Now(), 0
			} else if time.Since(lastChange) > 5*time.Minute {
				restarts++
				if restarts > 3 {
					log.Fatalf("No BLE advertisements for %s after %d scan restarts", time.Since(lastChange).Round(time.Second), restarts-1)
				}
				log.Printf("No BLE advertisements for 5 minutes, restarting the scan (%d)", restarts)
				tankSensors.StopScan()
				tankSensors.StartScan()
				lastChange = time.Now()
			}
			metrics.Update(tankSensors, cfg)
			pub.Publish(tankSensors)
		}
	}()

	srv := &http.Server{
		Handler:      api.NewAPIRouter(tankSensors, cfg),
		Addr:         cfg.ListenAddress,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdown)
	}()
	log.Printf("Listening on %s", cfg.ListenAddress)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
	pub.Close()
	tankSensors.StopScan()
	os.Exit(0)
}
