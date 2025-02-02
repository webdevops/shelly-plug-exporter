package shellyplug

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/webdevops/shelly-plug-exporter/discovery"
	"github.com/webdevops/shelly-plug-exporter/shellyprober"
)

type (
	shellyGen2ConfigValue struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}
)

func (sp *ShellyPlug) collectFromTargetGen2(target discovery.DiscoveryTarget, logger *zap.SugaredLogger, infoLabels, targetLabels prometheus.Labels) {
	sp.prometheus.info.With(infoLabels).Set(1)

	shellyProber := shellyprober.ShellyProberGen2{
		Target: target,
		Client: sp.client,
		Ctx:    sp.ctx,
		Cache:  globalCache,
	}

	if shellyConfig, err := shellyProber.GetShellyConfig(); err == nil {
		// systemStatus
		if result, err := shellyProber.GetSysStatus(); err == nil {
			sp.prometheus.sysUnixtime.With(targetLabels).Set(float64(result.Unixtime))
			sp.prometheus.sysUptime.With(targetLabels).Set(float64(result.Uptime))
			sp.prometheus.sysMemTotal.With(targetLabels).Set(float64(result.RAMSize))
			sp.prometheus.sysMemFree.With(targetLabels).Set(float64(result.RAMFree))
			sp.prometheus.sysFsSize.With(targetLabels).Set(float64(result.FsSize))
			sp.prometheus.sysFsFree.With(targetLabels).Set(float64(result.FsFree))
			sp.prometheus.restartRequired.With(targetLabels).Set(boolToFloat64(result.RestartRequired))

			if result.AvailableUpdates.Stable.Version != "" {
				sp.prometheus.updateNeeded.With(targetLabels).Set(1)
			} else {
				sp.prometheus.updateNeeded.With(targetLabels).Set(0)
			}
		} else {
			logger.Errorf(`failed to decode sysConfig: %v`, err)
		}

		// wifiStatus
		if result, err := shellyProber.GetWifiStatus(); err == nil {
			wifiLabels := copyLabelMap(targetLabels)
			wifiLabels["ssid"] = result.Ssid
			sp.prometheus.wifiRssi.With(wifiLabels).Set(float64(result.Rssi))
		} else {
			logger.Errorf(`failed to decode wifiStatus: %v`, err)
		}

		for configName, configValue := range shellyConfig {
			switch {
			// switch
			case strings.HasPrefix(configName, "switch:"):
				if configData, err := decodeShellyConfigValueToItem(configValue); err == nil {
					if result, err := shellyProber.GetSwitchStatus(configData.Id); err == nil {
						switchLabels := copyLabelMap(targetLabels)
						switchLabels["id"] = fmt.Sprintf("switch:%d", configData.Id)
						switchLabels["name"] = configData.Name

						switchOnLabels := copyLabelMap(switchLabels)
						switchOnLabels["source"] = result.Source

						sp.prometheus.switchOn.With(switchOnLabels).Set(boolToFloat64(result.Output))

						powerUsageLabels := copyLabelMap(targetLabels)
						powerUsageLabels["id"] = fmt.Sprintf("switch:%d", configData.Id)
						powerUsageLabels["name"] = configData.Name
						sp.prometheus.powerLoadCurrent.With(powerUsageLabels).Set(result.Apower)
						sp.prometheus.powerVoltage.With(powerUsageLabels).Set(result.Voltage)
						sp.prometheus.powerAmpere.With(powerUsageLabels).Set(result.Current)
					} else {
						logger.Errorf(`failed to decode switchStatus: %v`, err)
					}
				}
			// em
			case strings.HasPrefix(configName, "em:"):
				if configData, err := decodeShellyConfigValueToItem(configValue); err == nil {
					if result, err := shellyProber.GetEmStatus(configData.Id); err == nil {
						// phase A
						phase := "A"
						powerUsageLabels := copyLabelMap(targetLabels)
						powerUsageLabels["id"] = fmt.Sprintf("em:%d:%s", configData.Id, phase)
						powerUsageLabels["name"] = configData.Name
						sp.prometheus.powerLoadCurrent.With(powerUsageLabels).Set(result.AActPower)
						sp.prometheus.powerLoadApparentCurrent.With(powerUsageLabels).Set(result.AAprtPower)
						sp.prometheus.powerFactor.With(powerUsageLabels).Set(result.APf)
						sp.prometheus.powerFrequency.With(powerUsageLabels).Set(result.AFreq)
						sp.prometheus.powerVoltage.With(powerUsageLabels).Set(result.AVoltage)
						sp.prometheus.powerAmpere.With(powerUsageLabels).Set(result.ACurrent)

						// phase B
						phase = "B"
						powerUsageLabels = copyLabelMap(targetLabels)
						powerUsageLabels["id"] = fmt.Sprintf("em:%d:%s", configData.Id, phase)
						powerUsageLabels["name"] = configData.Name
						sp.prometheus.powerLoadCurrent.With(powerUsageLabels).Set(result.BActPower)
						sp.prometheus.powerLoadApparentCurrent.With(powerUsageLabels).Set(result.BAprtPower)
						sp.prometheus.powerFactor.With(powerUsageLabels).Set(result.BPf)
						sp.prometheus.powerFrequency.With(powerUsageLabels).Set(result.BFreq)
						sp.prometheus.powerVoltage.With(powerUsageLabels).Set(result.BVoltage)
						sp.prometheus.powerAmpere.With(powerUsageLabels).Set(result.BCurrent)

						// phase C
						phase = "C"
						powerUsageLabels = copyLabelMap(targetLabels)
						powerUsageLabels["id"] = fmt.Sprintf("em:%d:%s", configData.Id, phase)
						powerUsageLabels["name"] = configData.Name
						sp.prometheus.powerLoadCurrent.With(powerUsageLabels).Set(result.CActPower)
						sp.prometheus.powerLoadApparentCurrent.With(powerUsageLabels).Set(result.CAprtPower)
						sp.prometheus.powerFactor.With(powerUsageLabels).Set(result.CPf)
						sp.prometheus.powerFrequency.With(powerUsageLabels).Set(result.CFreq)
						sp.prometheus.powerVoltage.With(powerUsageLabels).Set(result.CVoltage)
						sp.prometheus.powerAmpere.With(powerUsageLabels).Set(result.CCurrent)
					} else {
						logger.Errorf(`failed to decode switchStatus: %v`, err)
					}

					if result, err := shellyProber.GetEmDataStatus(configData.Id); err == nil {
						// phase A
						phase := "A"
						powerUsageLabels := copyLabelMap(targetLabels)
						powerUsageLabels["id"] = fmt.Sprintf("em:%d:%s", configData.Id, phase)
						powerUsageLabels["name"] = configData.Name
						powerUsageLabels["direction"] = "in"
						sp.prometheus.powerLoadTotal.With(powerUsageLabels).Set(result.ATotalActEnergy)

						powerUsageLabels = copyLabelMap(targetLabels)
						powerUsageLabels["id"] = fmt.Sprintf("em:%d:%s", configData.Id, phase)
						powerUsageLabels["name"] = configData.Name
						powerUsageLabels["direction"] = "out"
						sp.prometheus.powerLoadTotal.With(powerUsageLabels).Set(result.ATotalActRetEnergy)

						// phase B
						phase = "B"
						powerUsageLabels = copyLabelMap(targetLabels)
						powerUsageLabels["id"] = fmt.Sprintf("em:%d:%s", configData.Id, phase)
						powerUsageLabels["name"] = configData.Name
						powerUsageLabels["direction"] = "in"
						sp.prometheus.powerLoadTotal.With(powerUsageLabels).Set(result.BTotalActEnergy)

						powerUsageLabels = copyLabelMap(targetLabels)
						powerUsageLabels["id"] = fmt.Sprintf("em:%d:%s", configData.Id, phase)
						powerUsageLabels["name"] = configData.Name
						powerUsageLabels["direction"] = "out"
						sp.prometheus.powerLoadTotal.With(powerUsageLabels).Set(result.BTotalActRetEnergy)

						// phase C
						phase = "C"
						powerUsageLabels = copyLabelMap(targetLabels)
						powerUsageLabels["id"] = fmt.Sprintf("em:%d:%s", configData.Id, phase)
						powerUsageLabels["name"] = configData.Name
						powerUsageLabels["direction"] = "in"
						sp.prometheus.powerLoadTotal.With(powerUsageLabels).Set(result.CTotalActEnergy)

						powerUsageLabels = copyLabelMap(targetLabels)
						powerUsageLabels["id"] = fmt.Sprintf("em:%d:%s", configData.Id, phase)
						powerUsageLabels["name"] = configData.Name
						powerUsageLabels["direction"] = "out"
						sp.prometheus.powerLoadTotal.With(powerUsageLabels).Set(result.CTotalActRetEnergy)
					} else {
						logger.Errorf(`failed to decode switchStatus: %v`, err)
					}
				}

			// temperatureSensor
			case strings.HasPrefix(configName, "temperature:"):
				if configData, err := decodeShellyConfigValueToItem(configValue); err == nil {
					if result, err := shellyProber.GetTemperatureStatus(configData.Id); err == nil {
						tempLabels := copyLabelMap(targetLabels)
						tempLabels["id"] = fmt.Sprintf("sensor:%d", configData.Id)
						tempLabels["name"] = configData.Name

						sp.prometheus.temp.With(tempLabels).Set(result.TC)
					} else {
						logger.Errorf(`failed to decode temperatureStatus: %v`, err)
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
