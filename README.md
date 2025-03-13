Shelly plug exporter
====================

[![license](https://img.shields.io/github/license/webdevops/shelly-plug-exporter.svg)](https://github.com/webdevops/shelly-plug-exporter/blob/master/LICENSE)
[![DockerHub](https://img.shields.io/badge/DockerHub-webdevops%2Fshelly--plug--exporter-blue)](https://hub.docker.com/r/webdevops/shelly-plug-exporter/)
[![Quay.io](https://img.shields.io/badge/Quay.io-webdevops%2Fshelly--plug--exporter-blue)](https://quay.io/repository/webdevops/shelly-plug-exporter)

Prometheus exporter for Shelly Plugs and devices (generation 1 and 2)
Can probe list of targets or use mDNS service discovery

Usage
-----

```
Usage:
  shelly-plug-exporter [OPTIONS]

Application Options:
      --debug                            debug mode [$DEBUG]
  -v, --verbose                          verbose mode [$VERBOSE]
      --log.json                         Switch log output to json format [$LOG_JSON]
      --shelly.request.timeout=          Request timeout (default: 5s) [$SHELLY_REQUEST_TIMEOUT]
      --shelly.auth.username=            Username for shelly plug login [$SHELLY_AUTH_USERNAME]
      --shelly.auth.password=            Password for shelly plug login [$SHELLY_AUTH_PASSWORD]
      --shelly.servicediscovery.timeout= mDNS discovery response timeout (default: 5s) [$SHELLY_SERVICEDISCOVERY_TIMEOUT]
      --shelly.servicediscovery.refresh= mDNS discovery refresh time (default: 15m) [$SHELLY_SERVICEDISCOVERY_REFRESH]
      --server.bind=                     Server address (default: :8080) [$SERVER_BIND]
      --server.timeout.read=             Server read timeout (default: 5s) [$SERVER_TIMEOUT_READ]
      --server.timeout.write=            Server write timeout (default: 10s) [$SERVER_TIMEOUT_WRITE]

Help Options:
  -h, --help                             Show this help message
```

Docker & Prometheus
-------------------

For mDNS service discovery the docker container must run on host network for multicast requests so the exporter needs
to be exposed on the host machine.

docker-compose.yaml
```yaml
version: '3.4'
services:
    # ...
    shelly-plug:
        image: webdevops/shelly-plug-exporter:$VERSION
        restart: always
        network_mode: host
        read_only: true
        environment:
            - VERBOSE=1
            - SERVER_BIND=:8089
            - SHELLY_SERVICEDISCOVERY_TIMEOUT=10s
    # ...
```

prometheus config:
```yaml
# ...
scrape_configs:
    # ...

    # general exporter metrics (eg memory & golang metrics)
    - job_name: 'shelly-plug-exporter'
      static_configs:
          - targets: ['host-addr:8089']

    # plugs metrics (mDNS servicediscovery)
    - job_name: 'shelly-plug-discovery'
      scrape_interval: 30s
      scrape_timeout: 29s
      static_configs:
          - targets: ['host-addr:8089']
      metrics_path: /probe/discovery

    # ...
```

HTTP Endpoints
--------------

| Endpoint   | Description                                                                                       |
|------------|---------------------------------------------------------------------------------------------------|
| `/metrics` | Default prometheus golang metrics                                                                 |
| `/probe`   | Probe shelly plugs, uses mDNS servicediscovery to find Shelly plugs (must be run on host network) |
| `/targets` | List of configured and discovered targets as JSON                                                 |

Metrics
-------

| Metric                                  | Description                                |
|-----------------------------------------|--------------------------------------------|
| `shellyplug_info`                       | Device information                         |
| `shellyplug_cloud_connected`            | Status if cloud connection established     |
| `shellyplug_cloud_enabled`              | Status if cloud connection enabled         |
| `shellyplug_overtemperature`            | Status if temperature reached limit        |
| `shellyplug_temperature`                | Device temperature                         |
| `shellyplug_switch_on`                  | Status if relay switch is on or off        |
| `shellyplug_switch_overpower`           | Status if relay switch triggered overpower |
| `shellyplug_switch_timer`               | Status if relay switch has timer           |
| `shellyplug_power_load_current`         | Current power load                         |
| `shellyplug_power_load_apparentcurrent` | Current power apparent load                |
| `shellyplug_power_load_total`           | Total power load in watt/hours             |
| `shellyplug_power_load_limit`           | Configured power limit                     |
| `shellyplug_power_factor`               | Power factor                               |
| `shellyplug_power_frequency`            | Power frequency in Hertz                   |
| `shellyplug_power_voltage`              | Power voltage                              |
| `shellyplug_power_ampere`               | Power ampere                               |
| `shellyplug_system_fs_free`             | System filesystem free space               |
| `shellyplug_system_fs_size`             | System filesystem size                     |
| `shellyplug_system_memory_free`         | System memory free                         |
| `shellyplug_system_memory_total`        | System memory size                         |
| `shellyplug_system_unixtime`            | System time (unixtime)                     |
| `shellyplug_system_uptime`              | System uptime (in seconds)                 |
| `shellyplug_update_needed`              | Status if updated is needed                |
| `shellyplug_restart_required`           | Status if restart of device is needed      |
| `shellyplug_wifi_rssi`                  | Wifi rssi                                  |
