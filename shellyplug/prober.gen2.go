package shellyplug

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/webdevops/shelly-plug-exporter/discovery"
	"github.com/webdevops/shelly-plug-exporter/shellyprober"
)

type (
	shellyGen2ConfigValue struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}
)

func (sp *ShellyPlug) collectFromTargetGen2(target discovery.DiscoveryTarget, logger *log.Entry, infoLabels, targetLabels prometheus.Labels) {
	sp.prometheus.info.With(infoLabels).Set(1)

	shellyProber := shellyprober.ShellyProberGen2{
		Target: target,
		Client: sp.client,
		Ctx:    sp.ctx,
	}

	if shellyConfig, err := shellyProber.GetShellyConfig(); err == nil {
		if result, err := shellyProber.GetSysStatus(); err == nil {
			sp.prometheus.sysUnixtime.With(targetLabels).Set(float64(result.Unixtime))
			sp.prometheus.sysUptime.With(targetLabels).Set(float64(result.Uptime))
			sp.prometheus.sysMemTotal.With(targetLabels).Set(float64(result.RAMSize))
			sp.prometheus.sysMemFree.With(targetLabels).Set(float64(result.RAMFree))
			sp.prometheus.sysFsSize.With(targetLabels).Set(float64(result.FsSize))
			sp.prometheus.sysFsFree.With(targetLabels).Set(float64(result.FsFree))
			sp.prometheus.restartRequired.With(targetLabels).Set(boolToFloat64(result.RestartRequired))
		} else {
			logger.Errorf(`failed to fetch status: %v`, err)
			if discovery.ServiceDiscovery != nil {
				discovery.ServiceDiscovery.MarkTarget(target.Address, discovery.TargetUnhealthy)
			}
		}

		if result, err := shellyProber.GetWifiStatus(); err == nil {
			wifiLabels := copyLabelMap(targetLabels)
			wifiLabels["ssid"] = result.Ssid
			sp.prometheus.wifiRssi.With(wifiLabels).Set(float64(result.Rssi))
		} else {
			logger.Errorf(`failed to fetch status: %v`, err)
			if discovery.ServiceDiscovery != nil {
				discovery.ServiceDiscovery.MarkTarget(target.Address, discovery.TargetUnhealthy)
			}
		}

		for configName, configValue := range shellyConfig {
			switch {
			case strings.HasPrefix(configName, "switch:"):
				if configData, err := decodeShellyConfigValueToItem(configValue); err == nil {
					if result, err := shellyProber.GetSwitchStatus(configData.Id); err == nil {
						switchLabels := copyLabelMap(targetLabels)
						switchLabels["switchID"] = strconv.Itoa(configData.Id)
						switchLabels["switchName"] = configData.Name

						switchOnLabels := copyLabelMap(switchLabels)
						switchOnLabels["switchSource"] = result.Source

						sp.prometheus.switchOn.With(switchOnLabels).Set(boolToFloat64(result.Output))

						powerUsageLabels := copyLabelMap(targetLabels)
						powerUsageLabels["switchID"] = strconv.Itoa(configData.Id)
						powerUsageLabels["switchName"] = configData.Name
						sp.prometheus.powerCurrent.With(powerUsageLabels).Set(result.Current)
						// total is provided as watt/minutes, we want watt/hours
						sp.prometheus.powerTotal.With(powerUsageLabels).Set(result.Apower / 60)
					} else {
						fmt.Println(err)
					}
				}
			case strings.HasPrefix(configName, "temperature:"):
				if configData, err := decodeShellyConfigValueToItem(configValue); err == nil {
					if result, err := shellyProber.GetTemperatureStatus(configData.Id); err == nil {
						tempLabels := copyLabelMap(targetLabels)
						tempLabels["sensorID"] = strconv.Itoa(configData.Id)
						tempLabels["sensorName"] = configData.Name

						sp.prometheus.temp.With(tempLabels).Set(result.TC)
					} else {
						logger.Errorf(`failed to fetch status: %v`, err)
						if discovery.ServiceDiscovery != nil {
							discovery.ServiceDiscovery.MarkTarget(target.Address, discovery.TargetUnhealthy)
						}
					}
				}
			}
		}

	} else {
		logger.Errorf(`failed to fetch status: %v`, err)
		if discovery.ServiceDiscovery != nil {
			discovery.ServiceDiscovery.MarkTarget(target.Address, discovery.TargetUnhealthy)
		}
	}
}

func decodeShellyConfigValueToItem(val interface{}) (shellyGen2ConfigValue, error) {
	ret := shellyGen2ConfigValue{}

	data, err := json.Marshal(val)
	if err != nil {
		return ret, err
	}

	err = json.Unmarshal(data, &ret)
	return ret, err
}
