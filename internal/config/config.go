// Package config reads pro-check-server settings from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jgulick48/mopeka_pro_check"
)

// Tank is the configuration of one sensor.
type Tank struct {
	Type string // key of mopeka_pro_check.TankTypes, e.g. "500g_h"
	Name string // display name, e.g. "Propane tank"
}

type Config struct {
	ListenAddress   string
	HCIDevice       int
	Tanks           map[string]Tank // by upper-case address
	DefaultTankType string
	StaleAfter      time.Duration

	MQTTBroker          string
	MQTTClientID        string
	MQTTUsername        string
	MQTTPassword        string
	MQTTCAFile          string
	MQTTCertFile        string
	MQTTKeyFile         string
	MQTTServerName      string
	MQTTTopicPrefix     string
	MQTTDiscoveryPrefix string
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}

// Load reads the configuration. TANKS is a comma separated list of
// ADDRESS=TYPE[:NAME], e.g. "E4:C4:15:4B:B6:9E=500g_h:Propane tank".
func Load() (Config, error) {
	c := Config{
		ListenAddress:       env("LISTEN_ADDRESS", ":8000"),
		Tanks:               map[string]Tank{},
		DefaultTankType:     env("DEFAULT_TANK_TYPE", ""),
		MQTTBroker:          env("MQTT_BROKER", ""),
		MQTTClientID:        env("MQTT_CLIENT_ID", "pro-check-server"),
		MQTTUsername:        env("MQTT_USERNAME", ""),
		MQTTPassword:        env("MQTT_PASSWORD", ""),
		MQTTCAFile:          env("MQTT_CA_FILE", ""),
		MQTTCertFile:        env("MQTT_CERT_FILE", ""),
		MQTTKeyFile:         env("MQTT_KEY_FILE", ""),
		MQTTServerName:      env("MQTT_SERVER_NAME", ""),
		MQTTTopicPrefix:     strings.TrimSuffix(env("MQTT_TOPIC_PREFIX", "mopeka"), "/"),
		MQTTDiscoveryPrefix: strings.TrimSuffix(env("MQTT_DISCOVERY_PREFIX", "homeassistant"), "/"),
	}
	var err error
	if c.HCIDevice, err = strconv.Atoi(env("HCI_DEVICE", "0")); err != nil {
		return c, fmt.Errorf("HCI_DEVICE: %w", err)
	}
	if c.StaleAfter, err = time.ParseDuration(env("STALE_AFTER", "1h")); err != nil {
		return c, fmt.Errorf("STALE_AFTER: %w", err)
	}
	if c.DefaultTankType != "" {
		if _, ok := mopeka_pro_check.TankTypes[c.DefaultTankType]; !ok {
			return c, fmt.Errorf("DEFAULT_TANK_TYPE %q is not a known tank type", c.DefaultTankType)
		}
	}
	for _, item := range strings.Split(env("TANKS", ""), ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		addr, rest, ok := strings.Cut(item, "=")
		if !ok {
			return c, fmt.Errorf("TANKS entry %q: want ADDRESS=TYPE[:NAME]", item)
		}
		typ, name, _ := strings.Cut(rest, ":")
		if _, ok := mopeka_pro_check.TankTypes[typ]; !ok {
			return c, fmt.Errorf("TANKS entry %q: unknown tank type %q", item, typ)
		}
		c.Tanks[strings.ToUpper(strings.TrimSpace(addr))] = Tank{Type: typ, Name: strings.TrimSpace(name)}
	}
	return c, nil
}

// TankFor is the configured tank for an address, falling back to the default type.
func (c Config) TankFor(addr string) Tank {
	if t, ok := c.Tanks[strings.ToUpper(addr)]; ok {
		return t
	}
	return Tank{Type: c.DefaultTankType}
}
