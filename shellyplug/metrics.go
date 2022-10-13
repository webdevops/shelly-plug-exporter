package shellyplug

import (
	"github.com/prometheus/client_golang/prometheus"
)

func (sp *ShellyPlug) initMetrics() {
	sp.prometheus.info = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_info",
			Help: "ShellyPlug info",
		},
		[]string{
			"target",
			"mac",
			"hostname",
			"plugName",
			"plugType",
			"plugSerial",
		},
	)
	sp.registry.MustRegister(sp.prometheus.info)

	sp.prometheus.temp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_temperature",
			Help: "ShellyPlug temperature",
		},
		[]string{"target", "mac", "plugName"},
	)
	sp.registry.MustRegister(sp.prometheus.temp)

	sp.prometheus.overTemp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_overtemperature",
			Help: "ShellyPlug over temperature",
		},
		[]string{"target", "mac", "plugName"},
	)
	sp.registry.MustRegister(sp.prometheus.overTemp)

	sp.prometheus.wifiRssi = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_wifi_rssi",
			Help: "ShellyPlug wifi rssi",
		},
		[]string{
			"target",
			"mac",
			"plugName",
			"ssid",
		},
	)
	sp.registry.MustRegister(sp.prometheus.wifiRssi)

	sp.prometheus.updateNeeded = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_update_needed",
			Help: "ShellyPlug status is update is needed",
		},
		[]string{"target", "mac", "plugName"},
	)
	sp.registry.MustRegister(sp.prometheus.updateNeeded)

	sp.prometheus.cloudEnabled = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_cloud_enabled",
			Help: "ShellyPlug status if cloud is enabled",
		},
		[]string{"target", "mac", "plugName"},
	)
	sp.registry.MustRegister(sp.prometheus.cloudEnabled)

	sp.prometheus.cloudConnected = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_cloud_connected",
			Help: "ShellyPlug status if device is connected to cloud",
		},
		[]string{"target", "mac", "plugName"},
	)
	sp.registry.MustRegister(sp.prometheus.cloudConnected)

	sp.prometheus.powerCurrent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_power_current",
			Help: "ShellyPlug current power usage in watts",
		},
		[]string{"target", "mac", "plugName"},
	)
	sp.registry.MustRegister(sp.prometheus.powerCurrent)

	sp.prometheus.powerTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_power_total",
			Help: "ShellyPlug current power total in watts",
		},
		[]string{"target", "mac", "plugName"},
	)
	sp.registry.MustRegister(sp.prometheus.powerTotal)

	sp.prometheus.powerLimit = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_power_limit",
			Help: "ShellyPlug configured power limit in watts",
		},
		[]string{"target", "mac", "plugName"},
	)
	sp.registry.MustRegister(sp.prometheus.powerLimit)

	sp.prometheus.sysUnixtime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_system_unixtime",
			Help: "ShellyPlug system unixtime",
		},
		[]string{"target", "mac", "plugName"},
	)
	sp.registry.MustRegister(sp.prometheus.sysUnixtime)

	sp.prometheus.sysUptime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_system_uptime",
			Help: "ShellyPlug system uptime",
		},
		[]string{"target", "mac", "plugName"},
	)
	sp.registry.MustRegister(sp.prometheus.sysUptime)

	sp.prometheus.sysMemTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_system_memory_total",
			Help: "ShellyPlug system memory total",
		},
		[]string{"target", "mac", "plugName"},
	)
	sp.registry.MustRegister(sp.prometheus.sysMemTotal)

	sp.prometheus.sysMemFree = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_system_memory_free",
			Help: "ShellyPlug system memory free",
		},
		[]string{"target", "mac", "plugName"},
	)
	sp.registry.MustRegister(sp.prometheus.sysMemFree)

	sp.prometheus.sysFsSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_system_fs_size",
			Help: "ShellyPlug system filesystem size",
		},
		[]string{"target", "mac", "plugName"},
	)
	sp.registry.MustRegister(sp.prometheus.sysFsSize)

	sp.prometheus.sysFsFree = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_system_fs_free",
			Help: "ShellyPlug system filesystem free",
		},
		[]string{"target", "mac", "plugName"},
	)
	sp.registry.MustRegister(sp.prometheus.sysFsFree)

}
