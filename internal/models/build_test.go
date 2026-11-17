package models

import (
	"encoding/hex"
	"math"
	"testing"

	"github.com/jgulick48/mopeka_pro_check"
)

func TestFromDevice(t *testing.T) {
	data, _ := hex.DecodeString("5900086e3738c64bb69e50e2")
	d, ok := mopeka_pro_check.NewFromManufacturerData("E4:C4:15:4B:B6:9E", "", -79, data)
	if !ok {
		t.Fatal("sample did not decode")
	}
	s := FromDevice(d, "500g_h", "Propane tank")
	if s.LevelPercent == nil || math.Abs(*s.LevelPercent-72.5) > 0.5 {
		t.Fatalf("levelPercent = %v", s.LevelPercent)
	}
	if s.Gallons == nil || math.Abs(*s.Gallons-362) > 3 {
		t.Fatalf("gallons = %v", s.Gallons)
	}
	if s.Name != "Propane tank" || s.RSSI != -79 {
		t.Fatalf("sensor = %+v", s)
	}
	if u := FromDevice(d, "", ""); u.LevelPercent != nil || u.Gallons != nil {
		t.Fatal("no tank type should leave percent and gallons unset")
	}
}
