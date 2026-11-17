# pro-check-server
An HTTP web server to expose Mopeka Pro Check sensor data.

## Running

The server scans BLE over a raw HCI socket, so it runs as root (or with
`CAP_NET_RAW` and `CAP_NET_ADMIN`) with host networking. BlueZ isn't needed.

Endpoints on `LISTEN_ADDRESS`:

- `/sensors`: JSON for every Mopeka sensor heard.
- `/metrics`: Prometheus gauges (`mopeka_level_percent`, `mopeka_level_gallons`,
  `mopeka_level_mm`, `mopeka_temperature_celsius`, `mopeka_battery_volts`,
  `mopeka_read_quality`, `mopeka_rssi_dbm`, `mopeka_last_seen_timestamp_seconds`,
  `mopeka_scanner_advertisements_total`, `mopeka_scanner_running`).
- `/healthz`: 200 while the scanner runs.

## Configuration (environment)

| Variable | Default | |
|---|---|---|
| `LISTEN_ADDRESS` | `:8000` | HTTP listen address |
| `HCI_DEVICE` | `0` | Bluetooth adapter (hciN) |
| `TANKS` | | `ADDRESS=TYPE[:NAME],...`, e.g. `E4:C4:15:4B:B6:9E=500g_h:Propane tank` |
| `DEFAULT_TANK_TYPE` | | tank type for sensors not in `TANKS` |
| `STALE_AFTER` | `1h` | Home Assistant marks a sensor unavailable after this long without a reading |
| `MQTT_BROKER` | | e.g. `ssl://mqtt.example.com:8883`; MQTT is off when empty |
| `MQTT_CLIENT_ID` | `pro-check-server` | |
| `MQTT_USERNAME`, `MQTT_PASSWORD` | | optional |
| `MQTT_CA_FILE` | | extra CA (added to the system roots) |
| `MQTT_CERT_FILE`, `MQTT_KEY_FILE` | | client certificate for mutual TLS |
| `MQTT_SERVER_NAME` | | TLS server name override |
| `MQTT_TOPIC_PREFIX` | `mopeka` | state on `<prefix>/mopeka_<address>/state`, availability on `<prefix>/status` |
| `MQTT_DISCOVERY_PREFIX` | `homeassistant` | Home Assistant MQTT discovery |

Tank types: `20lb_v`, `30lb_v`, `40lb_v`, `100lb_v` (vertical) and `500g_h`
(500 gal horizontal; also reports gallons).

## Building

```
docker buildx build --platform linux/amd64,linux/arm64,linux/arm/v7 -t jgulick48/pro-check-server:<tag> --push .
```
