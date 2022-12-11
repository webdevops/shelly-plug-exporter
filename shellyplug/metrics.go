package shellyplug

import (
	"github.com/prometheus/client_golang/prometheus"
)

func (sp *ShellyPlug) initMetrics() {
	commonLabels := []string{"target", "mac", "plugName"}
	switchLabels := append(commonLabels, "unit")
	powerLabels := append(commonLabels, "unit")

	// ##########################################
	// Info

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
		},
	)
	sp.registry.MustRegister(sp.prometheus.info)

	// ##########################################
	// Temp

	sp.prometheus.temp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_temperature",
			Help: "ShellyPlug temperature",
		},
		commonLabels,
	)
	sp.registry.MustRegister(sp.prometheus.temp)

	sp.prometheus.overTemp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_overtemperature",
			Help: "ShellyPlug over temperature",
		},
		commonLabels,
	)
	sp.registry.MustRegister(sp.prometheus.overTemp)

	// ##########################################
	// Wifi

	sp.prometheus.wifiRssi = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_wifi_rssi",
			Help: "ShellyPlug wifi rssi",
		},
		[]string{"target", "mac", "plugName", "ssid"},
	)
	sp.registry.MustRegister(sp.prometheus.wifiRssi)

	// ##########################################
	// Update

	sp.prometheus.updateNeeded = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_update_needed",
			Help: "ShellyPlug status is update is needed",
		},
		commonLabels,
	)
	sp.registry.MustRegister(sp.prometheus.updateNeeded)

	// ##########################################
	// Cloud

	sp.prometheus.cloudEnabled = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_cloud_enabled",
			Help: "ShellyPlug status if cloud is enabled",
		},
		commonLabels,
	)
	sp.registry.MustRegister(sp.prometheus.cloudEnabled)

	sp.prometheus.cloudConnected = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_cloud_connected",
			Help: "ShellyPlug status if device is connected to cloud",
		},
		commonLabels,
	)
	sp.registry.MustRegister(sp.prometheus.cloudConnected)

	// ##########################################
	// Switch

	sp.prometheus.switchOn = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_switch_on",
			Help: "ShellyPlug switch on status",
		},
		append(switchLabels, "switchSource"),
	)
	sp.registry.MustRegister(sp.prometheus.switchOn)

	sp.prometheus.switchOverpower = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_switch_overpower",
			Help: "ShellyPlug switch overpower status",
		},
		switchLabels,
	)
	sp.registry.MustRegister(sp.prometheus.switchOverpower)

	sp.prometheus.switchTimer = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_switch_timer",
			Help: "ShellyPlug status if time is active",
		},
		switchLabels,
	)
	sp.registry.MustRegister(sp.prometheus.switchTimer)

	// ##########################################
	// Power

	sp.prometheus.powerCurrent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_power_current",
			Help: "ShellyPlug current power usage in watts",
		},
		powerLabels,
	)
	sp.registry.MustRegister(sp.prometheus.powerCurrent)

	sp.prometheus.powerTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_power_total",
			Help: "ShellyPlug current power total in watts",
		},
		powerLabels,
	)
	sp.registry.MustRegister(sp.prometheus.powerTotal)

	sp.prometheus.powerLimit = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_power_limit",
			Help: "ShellyPlug configured power limit in watts",
		},
		commonLabels,
	)
	sp.registry.MustRegister(sp.prometheus.powerLimit)

	// ##########################################
	// System

	sp.prometheus.sysUnixtime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_system_unixtime",
			Help: "ShellyPlug system unixtime",
		},
		commonLabels,
	)
	sp.registry.MustRegister(sp.prometheus.sysUnixtime)

	sp.prometheus.sysUptime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_system_uptime",
			Help: "ShellyPlug system uptime",
		},
		commonLabels,
	)
	sp.registry.MustRegister(sp.prometheus.sysUptime)

	sp.prometheus.sysMemTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_system_memory_total",
			Help: "ShellyPlug system memory total",
		},
		commonLabels,
	)
	sp.registry.MustRegister(sp.prometheus.sysMemTotal)

	sp.prometheus.sysMemFree = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_system_memory_free",
			Help: "ShellyPlug system memory free",
		},
		commonLabels,
	)
	sp.registry.MustRegister(sp.prometheus.sysMemFree)

	sp.prometheus.sysFsSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_system_fs_size",
			Help: "ShellyPlug system filesystem size",
		},
		commonLabels,
	)
	sp.registry.MustRegister(sp.prometheus.sysFsSize)

	sp.prometheus.sysFsFree = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_system_fs_free",
			Help: "ShellyPlug system filesystem free",
		},
		commonLabels,
	)
	sp.registry.MustRegister(sp.prometheus.sysFsFree)

}
