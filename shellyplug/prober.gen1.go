package shellyplug

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/webdevops/shelly-plug-exporter/discovery"
	"github.com/webdevops/shelly-plug-exporter/shellyprober"
)

func (sp *ShellyPlug) collectFromTargetGen1(target discovery.DiscoveryTarget, logger *log.Entry, infoLabels, targetLabels prometheus.Labels) {
	shellyProber := shellyprober.ShellyProberGen1{
		Target: target,
		Client: sp.client,
		Ctx:    sp.ctx,
	}

	if result, err := shellyProber.GetSettings(); err == nil {
		if discovery.ServiceDiscovery != nil {
			discovery.ServiceDiscovery.MarkTarget(target.Address, discovery.DiscoveryTargetHealthy)
		}

		targetLabels["plugName"] = result.Name

		infoLabels["plugName"] = result.Name
		infoLabels["plugType"] = result.Device.Type

		powerLimitLabels := copyLabelMap(targetLabels)
		powerLimitLabels["switchID"] = ""
		powerLimitLabels["switchName"] = ""
		sp.prometheus.powerLimit.With(powerLimitLabels).Set(result.MaxPower)
	} else {
		logger.Errorf(`failed to fetch settings: %v`, err)
		if discovery.ServiceDiscovery != nil {
			discovery.ServiceDiscovery.MarkTarget(target.Address, discovery.DiscoveryTargetUnhealthy)
		}
	}

	sp.prometheus.info.With(infoLabels).Set(1)

	if result, err := shellyProber.GetStatus(); err == nil {
		sp.prometheus.sysUnixtime.With(targetLabels).Set(float64(result.Unixtime))
		sp.prometheus.sysUptime.With(targetLabels).Set(float64(result.Uptime))
		sp.prometheus.sysMemTotal.With(targetLabels).Set(float64(result.RAMTotal))
		sp.prometheus.sysMemFree.With(targetLabels).Set(float64(result.RAMFree))
		sp.prometheus.sysFsSize.With(targetLabels).Set(float64(result.FsSize))
		sp.prometheus.sysFsFree.With(targetLabels).Set(float64(result.FsFree))

		tempLabels := copyLabelMap(targetLabels)
		tempLabels["sensorID"] = ""
		tempLabels["sensorName"] = "system"
		sp.prometheus.temp.With(tempLabels).Set(result.Temperature)
		sp.prometheus.overTemp.With(tempLabels).Set(boolToFloat64(result.Overtemperature))

		wifiLabels := copyLabelMap(targetLabels)
		wifiLabels["ssid"] = result.WifiSta.Ssid
		sp.prometheus.wifiRssi.With(wifiLabels).Set(float64(result.WifiSta.Rssi))

		sp.prometheus.updateNeeded.With(targetLabels).Set(boolToFloat64(result.HasUpdate))
		sp.prometheus.cloudEnabled.With(targetLabels).Set(boolToFloat64(result.Cloud.Enabled))
		sp.prometheus.cloudConnected.With(targetLabels).Set(boolToFloat64(result.Cloud.Connected))

		for relayID, powerUsage := range result.Meters {
			powerUsageLabels := copyLabelMap(targetLabels)
			powerUsageLabels["switchID"] = strconv.Itoa(relayID)
			powerUsageLabels["switchName"] = targetLabels["plugName"]

			sp.prometheus.powerCurrent.With(powerUsageLabels).Set(powerUsage.Power)
			// total is provided as watt/minutes, we want watt/hours
			sp.prometheus.powerTotal.With(powerUsageLabels).Set(powerUsage.Total / 60)
		}

		for relayID, relay := range result.Relays {
			switchLabels := copyLabelMap(targetLabels)
			switchLabels["switchID"] = strconv.Itoa(relayID)
			switchLabels["switchName"] = targetLabels["plugName"]

			switchOnLabels := copyLabelMap(switchLabels)
			switchOnLabels["switchSource"] = relay.Source

			sp.prometheus.switchOn.With(switchOnLabels).Set(boolToFloat64(relay.Ison))
			sp.prometheus.switchOverpower.With(switchLabels).Set(boolToFloat64(relay.Overpower))
			sp.prometheus.switchTimer.With(switchLabels).Set(boolToFloat64(relay.HasTimer))
		}
	} else {
		logger.Errorf(`failed to fetch status: %v`, err)
		if discovery.ServiceDiscovery != nil {
			discovery.ServiceDiscovery.MarkTarget(target.Address, discovery.DiscoveryTargetUnhealthy)
		}
	}
}
