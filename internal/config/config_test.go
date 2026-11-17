package config

import (
	"testing"
	"time"
)

func TestLoadTanks(t *testing.T) {
	t.Setenv("TANKS", "e4:c4:15:4b:b6:9e=500g_h:Propane tank, AA:BB:CC:DD:EE:FF=20lb_v")
	t.Setenv("STALE_AFTER", "30m")
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got := c.TankFor("E4:C4:15:4B:B6:9E"); got.Type != "500g_h" || got.Name != "Propane tank" {
		t.Fatalf("tank = %+v", got)
	}
	if got := c.TankFor("aa:bb:cc:dd:ee:ff"); got.Type != "20lb_v" || got.Name != "" {
		t.Fatalf("tank = %+v", got)
	}
	if got := c.TankFor("11:22:33:44:55:66"); got.Type != "" {
		t.Fatalf("unconfigured tank = %+v", got)
	}
	if c.StaleAfter != 30*time.Minute || c.ListenAddress != ":8000" {
		t.Fatalf("defaults: %+v", c)
	}
}

func TestLoadRejectsUnknownType(t *testing.T) {
	t.Setenv("TANKS", "E4:C4:15:4B:B6:9E=1000g_h")
	if _, err := Load(); err == nil {
		t.Fatal("expected an error for an unknown tank type")
	}
}
