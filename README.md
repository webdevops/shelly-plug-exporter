Shelly plug exporter
====================

[![license](https://img.shields.io/github/license/webdevops/shelly-plug-exporter.svg)](https://github.com/webdevops/shelly-plug-exporter/blob/master/LICENSE)
[![DockerHub](https://img.shields.io/badge/DockerHub-webdevops%2Fshelly--plug--exporter-blue)](https://hub.docker.com/r/webdevops/shelly-plug-exporter/)
[![Quay.io](https://img.shields.io/badge/Quay.io-webdevops%2Fshelly--plug--exporter-blue)](https://quay.io/repository/webdevops/shelly-plug-exporter)

Prometheus exporter for Shelly Plugs

Usage
-----

```
Usage:
  shelly-plug-exporter [OPTIONS]

Application Options:
      --debug                 debug mode [$DEBUG]
  -v, --verbose               verbose mode [$VERBOSE]
      --log.json              Switch log output to json format [$LOG_JSON]
      --shelly.auth.username= Username for shelly plug login [$SHELLY_AUTH_USERNAME]
      --shelly.auth.password= Password for shelly plug login [$SHELLY_AUTH_PASSWORD]
      --server.bind=          Server address (default: :8080) [$SERVER_BIND]
      --server.timeout.read=  Server read timeout (default: 5s) [$SERVER_TIMEOUT_READ]
      --server.timeout.write= Server write timeout (default: 10s) [$SERVER_TIMEOUT_WRITE]

Help Options:
  -h, --help                  Show this help message
```

HTTP Endpoints
--------------

| Endpoint   | Description                                                                                 |
|------------|---------------------------------------------------------------------------------------------|
| `/metrics` | Default prometheus golang metrics                                                           |
| `/probe`   | Probe shelly plugs (with `?target=xxx.xxx.xxx.xxx` endpoints, can probe multiple endpoints) |

Metrics
-------

| Metric                           | Description                            |
|----------------------------------|----------------------------------------|
| `shellyplug_info`                | Device information                     |
| `shellyplug_cloud_connected`     | Status if cloud connection established |
| `shellyplug_cloud_enabled`       | Status if cloud connection enabled     |
| `shellyplug_overtemperature`     | Status if temperature reached limit    |
| `shellyplug_temperature`         | Device temperature                     |
| `shellyplug_power_current`       | Current power usage                    |
| `shellyplug_power_total`         | Total power usage                      |
| `shellyplug_power_limit`         | Configured power limit                 |
| `shellyplug_system_fs_free`      | System filesystem free space           |
| `shellyplug_system_fs_size`      | System filesystem size                 |
| `shellyplug_system_memory_free`  | System memory free                     |
| `shellyplug_system_memory_total` | System memory size                     |
| `shellyplug_system_unixtime`     | System time (unixtime)                 |
| `shellyplug_system_uptime`       | System uptime (in seconds)             |
| `shellyplug_update_needed`       | Status if updated is needed            |
| `shellyplug_wifi_rssi`           | Wifi rssi                              |
