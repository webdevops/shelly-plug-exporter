package shellyplug

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/webdevops/shelly-plug-exporter/discovery"
	"github.com/webdevops/shelly-plug-exporter/shellyprober"
)

func (sp *ShellyPlug) collectFromTargetGen1em(target discovery.DiscoveryTarget, logger *log.Entry, infoLabels, targetLabels prometheus.Labels) {
	shellyProber := shellyprober.ShellyProberGen1Em{
		Target: target,
		Client: sp.client,
		Ctx:    sp.ctx,
		Cache:  globalCache,
	}

	if result, err := shellyProber.GetSettings(); err == nil {
		if discovery.ServiceDiscovery != nil {
			discovery.ServiceDiscovery.MarkTarget(target.Address, discovery.TargetHealthy)
		}

		targetLabels["plugName"] = result.Name

		infoLabels["plugName"] = result.Name
		infoLabels["plugModel"] = result.Device.Type

		powerLimitLabels := copyLabelMap(targetLabels)
		powerLimitLabels["id"] = "emeter:0"
		powerLimitLabels["name"] = ""
		sp.prometheus.powerLoadLimit.With(powerLimitLabels).Set(result.MaxPower)
	} else {
		logger.Errorf(`failed to fetch settings: %v`, err)
		if discovery.ServiceDiscovery != nil {
			discovery.ServiceDiscovery.MarkTarget(target.Address, discovery.TargetUnhealthy)
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

		wifiLabels := copyLabelMap(targetLabels)
		wifiLabels["ssid"] = result.WifiSta.Ssid
		sp.prometheus.wifiRssi.With(wifiLabels).Set(float64(result.WifiSta.Rssi))

		sp.prometheus.updateNeeded.With(targetLabels).Set(boolToFloat64(result.HasUpdate))
		sp.prometheus.cloudEnabled.With(targetLabels).Set(boolToFloat64(result.Cloud.Enabled))
		sp.prometheus.cloudConnected.With(targetLabels).Set(boolToFloat64(result.Cloud.Connected))

		for relayID, powerUsage := range result.Emeters {
			powerUsageLabels := copyLabelMap(targetLabels)
			powerUsageLabels["id"] = fmt.Sprintf("emeter:%d", relayID)
			powerUsageLabels["name"] = targetLabels["plugName"]

			sp.prometheus.powerLoadCurrent.With(powerUsageLabels).Set(powerUsage.Power * -1)
			// total is provided as watt/minutes, we want watt/hours
			powerUsageLabels["direction"] = "in"
			sp.prometheus.powerLoadTotal.With(powerUsageLabels).Set(powerUsage.Total / 60)
		}

		for relayID, relay := range result.Relays {
			switchLabels := copyLabelMap(targetLabels)
			switchLabels["id"] = fmt.Sprintf("relay:%d", relayID)
			switchLabels["name"] = targetLabels["plugName"]

			switchOnLabels := copyLabelMap(switchLabels)
			switchOnLabels["source"] = relay.Source

			sp.prometheus.switchOn.With(switchOnLabels).Set(boolToFloat64(relay.Ison))
			sp.prometheus.switchOverpower.With(switchLabels).Set(boolToFloat64(relay.Overpower))
			sp.prometheus.switchTimer.With(switchLabels).Set(boolToFloat64(relay.HasTimer))
		}
	} else {
		logger.Errorf(`failed to fetch status: %v`, err)
		if discovery.ServiceDiscovery != nil {
			discovery.ServiceDiscovery.MarkTarget(target.Address, discovery.TargetUnhealthy)
		}
	}
}