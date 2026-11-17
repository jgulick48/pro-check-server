// Package mqttpub publishes readings to MQTT with Home Assistant discovery.
package mqttpub

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/jgulick48/mopeka_pro_check"
	"github.com/jgulick48/pro-check-server/internal/config"
	"github.com/jgulick48/pro-check-server/internal/models"
)

type Publisher struct {
	cfg        config.Config
	client     mqtt.Client
	mu         sync.Mutex
	published  map[string]time.Time // last reading time published, by address
	discovered map[string]bool
}

func tlsConfig(c config.Config) (*tls.Config, error) {
	t := &tls.Config{MinVersion: tls.VersionTLS12, ServerName: c.MQTTServerName}
	if c.MQTTCAFile != "" {
		pool, err := x509.SystemCertPool()
		if err != nil || pool == nil {
			pool = x509.NewCertPool()
		}
		pem, err := os.ReadFile(c.MQTTCAFile)
		if err != nil {
			return nil, fmt.Errorf("MQTT_CA_FILE: %w", err)
		}
		pool.AppendCertsFromPEM(pem)
		t.RootCAs = pool
	}
	if c.MQTTCertFile != "" || c.MQTTKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(c.MQTTCertFile, c.MQTTKeyFile)
		if err != nil {
			return nil, fmt.Errorf("MQTT client certificate: %w", err)
		}
		t.Certificates = []tls.Certificate{cert}
	}
	return t, nil
}

// New connects in the background; it returns nil if MQTT_BROKER is unset.
func New(c config.Config) (*Publisher, error) {
	if c.MQTTBroker == "" {
		return nil, nil
	}
	p := &Publisher{cfg: c, published: map[string]time.Time{}, discovered: map[string]bool{}}
	opts := mqtt.NewClientOptions().AddBroker(c.MQTTBroker).SetClientID(c.MQTTClientID).
		SetAutoReconnect(true).SetConnectRetry(true).SetConnectRetryInterval(10*time.Second).
		SetMaxReconnectInterval(time.Minute).SetKeepAlive(30*time.Second).
		SetWill(c.MQTTTopicPrefix+"/status", "offline", 1, true)
	if c.MQTTUsername != "" {
		opts.SetUsername(c.MQTTUsername).SetPassword(c.MQTTPassword)
	}
	if strings.HasPrefix(c.MQTTBroker, "ssl://") || strings.HasPrefix(c.MQTTBroker, "tls://") || strings.HasPrefix(c.MQTTBroker, "mqtts://") {
		t, err := tlsConfig(c)
		if err != nil {
			return nil, err
		}
		opts.SetTLSConfig(t)
	}
	opts.SetOnConnectHandler(func(cl mqtt.Client) {
		log.Printf("MQTT connected to %s", c.MQTTBroker)
		cl.Publish(c.MQTTTopicPrefix+"/status", 1, true, "online")
		p.mu.Lock()
		p.discovered = map[string]bool{} // re-announce after (re)connect
		p.published = map[string]time.Time{}
		p.mu.Unlock()
	})
	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) { log.Printf("MQTT connection lost: %v", err) })
	p.client = mqtt.NewClient(opts)
	p.client.Connect() // retries in the background
	return p, nil
}

func objectID(addr string) string {
	return "mopeka_" + strings.ToLower(strings.ReplaceAll(addr, ":", ""))
}

func (p *Publisher) stateTopic(addr string) string {
	return fmt.Sprintf("%s/%s/state", p.cfg.MQTTTopicPrefix, objectID(addr))
}

type entity struct {
	key, name, unit, deviceClass, stateClass, icon, category, template string
}

func (p *Publisher) discover(s models.Sensor) {
	id := objectID(s.Address)
	name := s.Name
	if name == "" {
		name = "Mopeka " + strings.ReplaceAll(s.Address[len(s.Address)-8:], ":", "")
	}
	device := map[string]any{
		"identifiers":  []string{id},
		"name":         name,
		"manufacturer": "Mopeka",
		"model":        "Pro Check (" + s.SensorType + ")",
		"connections":  [][]string{{"bluetooth", s.Address}},
	}
	ents := []entity{
		{"level_in", "Level", "in", "distance", "measurement", "mdi:propane-tank", "", "{{ value_json.tankLevelInches | round(1) }}"},
		{"temperature", "Temperature", "°F", "temperature", "measurement", "", "", "{{ value_json.tempFahrenheit | round(0) }}"},
		{"battery", "Battery", "%", "battery", "measurement", "", "diagnostic", "{{ value_json.batteryLevel }}"},
		{"battery_voltage", "Battery voltage", "V", "voltage", "measurement", "", "diagnostic", "{{ value_json.batteryVoltage | round(2) }}"},
		{"read_quality", "Read quality", "", "", "measurement", "mdi:signal", "diagnostic", "{{ value_json.sensorReadQuality }}"},
		{"rssi", "Signal strength", "dBm", "signal_strength", "measurement", "", "diagnostic", "{{ value_json.rssi }}"},
	}
	if s.LevelPercent != nil {
		ents = append(ents, entity{"level_percent", "Fill", "%", "", "measurement", "mdi:propane-tank", "", "{{ value_json.levelPercent | round(1) }}"})
	}
	if s.Gallons != nil {
		ents = append(ents, entity{"gallons", "Propane", "gal", "volume_storage", "measurement", "mdi:propane-tank", "", "{{ value_json.gallons | round(0) }}"})
	}
	for _, e := range ents {
		cfg := map[string]any{
			"name":               e.name,
			"unique_id":          id + "_" + e.key,
			"default_entity_id":  "sensor." + id + "_" + e.key,
			"state_topic":        p.stateTopic(s.Address),
			"value_template":     e.template,
			"availability_topic": p.cfg.MQTTTopicPrefix + "/status",
			"expire_after":       int(p.cfg.StaleAfter.Seconds()),
			"device":             device,
		}
		if e.unit != "" {
			cfg["unit_of_measurement"] = e.unit
		}
		if e.deviceClass != "" {
			cfg["device_class"] = e.deviceClass
		}
		if e.stateClass != "" {
			cfg["state_class"] = e.stateClass
		}
		if e.icon != "" {
			cfg["icon"] = e.icon
		}
		if e.category != "" {
			cfg["entity_category"] = e.category
		}
		b, _ := json.Marshal(cfg)
		p.client.Publish(fmt.Sprintf("%s/sensor/%s/%s/config", p.cfg.MQTTDiscoveryPrefix, id, e.key), 1, true, b)
	}
}

// Publish sends discovery (once per connection) and the state of every sensor
// that has a new reading.
func (p *Publisher) Publish(s *mopeka_pro_check.Scanner) {
	if p == nil || !p.client.IsConnectionOpen() {
		return
	}
	for _, d := range s.GetDevices() {
		tank := p.cfg.TankFor(d.GetAddress())
		view := models.FromDevice(d, tank.Type, tank.Name)
		p.mu.Lock()
		if !p.discovered[view.Address] {
			p.discover(view)
			p.discovered[view.Address] = true
		}
		fresh := d.GetReadingTime().After(p.published[view.Address])
		if fresh {
			p.published[view.Address] = d.GetReadingTime()
		}
		p.mu.Unlock()
		if fresh {
			b, _ := json.Marshal(view)
			p.client.Publish(p.stateTopic(view.Address), 1, true, b)
		}
	}
}

func (p *Publisher) Close() {
	if p == nil {
		return
	}
	p.client.Publish(p.cfg.MQTTTopicPrefix+"/status", 1, true, "offline").WaitTimeout(2 * time.Second)
	p.client.Disconnect(500)
}
