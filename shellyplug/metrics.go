package shellyplug

import (
	"github.com/prometheus/client_golang/prometheus"
)

type (
	shellyPlugMetrics struct {
		info            *prometheus.GaugeVec
		temp            *prometheus.GaugeVec
		overTemp        *prometheus.GaugeVec
		wifiRssi        *prometheus.GaugeVec
		updateNeeded    *prometheus.GaugeVec
		restartRequired *prometheus.GaugeVec

		cloudEnabled   *prometheus.GaugeVec
		cloudConnected *prometheus.GaugeVec

		switchOn        *prometheus.GaugeVec
		switchOverpower *prometheus.GaugeVec
		switchTimer     *prometheus.GaugeVec

		powerCurrent         *prometheus.GaugeVec
		powerApparentCurrent *prometheus.GaugeVec
		powerTotal           *prometheus.GaugeVec
		powerLimit           *prometheus.GaugeVec
		powerFactor          *prometheus.GaugeVec
		powerFrequency       *prometheus.GaugeVec
		powerVoltage         *prometheus.GaugeVec
		powerAmpere          *prometheus.GaugeVec

		sysUnixtime *prometheus.GaugeVec
		sysUptime   *prometheus.GaugeVec
		sysMemTotal *prometheus.GaugeVec
		sysMemFree  *prometheus.GaugeVec
		sysFsSize   *prometheus.GaugeVec
		sysFsFree   *prometheus.GaugeVec
	}
)

func (sp *ShellyPlug) initMetrics() {
	commonLabels := []string{"target", "mac", "plugName"}
	tempLabels := append(commonLabels, "id", "name")
	switchLabels := append(commonLabels, "id", "name")
	powerLabels := append(commonLabels, "id", "name")

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
			"plugModel",
			"plugApp",
			"plugGeneration",
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
		tempLabels,
	)
	sp.registry.MustRegister(sp.prometheus.temp)

	sp.prometheus.overTemp = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_overtemperature",
			Help: "ShellyPlug over temperature",
		},
		tempLabels,
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

	sp.prometheus.restartRequired = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_restart_required",
			Help: "ShellyPlug if restart is required",
		},
		commonLabels,
	)
	sp.registry.MustRegister(sp.prometheus.restartRequired)

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
		append(switchLabels, "source"),
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
			Help: "ShellyPlug current power current in watts",
		},
		powerLabels,
	)
	sp.registry.MustRegister(sp.prometheus.powerCurrent)

	sp.prometheus.powerApparentCurrent = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_power_apparentcurrent",
			Help: "ShellyPlug current power apparent current in VA",
		},
		powerLabels,
	)
	sp.registry.MustRegister(sp.prometheus.powerApparentCurrent)

	sp.prometheus.powerTotal = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_power_total",
			Help: "ShellyPlug current power total in watts",
		},
		append(powerLabels, "direction"),
	)
	sp.registry.MustRegister(sp.prometheus.powerTotal)

	sp.prometheus.powerLimit = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_power_limit",
			Help: "ShellyPlug configured power limit in watts",
		},
		powerLabels,
	)
	sp.registry.MustRegister(sp.prometheus.powerLimit)

	sp.prometheus.powerFactor = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_power_factor",
			Help: "ShellyPlug configured power factor",
		},
		powerLabels,
	)
	sp.registry.MustRegister(sp.prometheus.powerFactor)

	sp.prometheus.powerFrequency = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_power_frequency",
			Help: "ShellyPlug configured power frequency in Hz",
		},
		powerLabels,
	)
	sp.registry.MustRegister(sp.prometheus.powerFrequency)

	sp.prometheus.powerVoltage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_power_voltage",
			Help: "ShellyPlug configured power voltage",
		},
		powerLabels,
	)
	sp.registry.MustRegister(sp.prometheus.powerVoltage)

	sp.prometheus.powerAmpere = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_power_ampere",
			Help: "ShellyPlug configured power ampere",
		},
		powerLabels,
	)
	sp.registry.MustRegister(sp.prometheus.powerAmpere)

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
