package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/webdevops/shelly-plug-exporter/shellyplug"
)

const (
	DefaultTimeout = 30
)

func shellyProbe(w http.ResponseWriter, r *http.Request) {
	var err error
	var timeoutSeconds float64

	// startTime := time.Now()
	contextLogger := buildContextLoggerFromRequest(r)
	registry := prometheus.NewRegistry()

	// If a timeout is configured via the Prometheus header, add it to the request.
	timeoutSeconds, err = getPrometheusTimeout(r, DefaultTimeout)
	if err != nil {
		contextLogger.Error(err)
		http.Error(w, fmt.Sprintf("failed to parse timeout from Prometheus header: %s", err), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds*float64(time.Second)))
	defer cancel()
	r = r.WithContext(ctx)

	metricInfo := prometheus.NewGaugeVec(
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
	registry.MustRegister(metricInfo)

	metricTemp := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_temperature",
			Help: "ShellyPlug temperature",
		},
		[]string{"target", "mac", "plugName"},
	)
	registry.MustRegister(metricTemp)

	metricOverTemp := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_overtemperature",
			Help: "ShellyPlug over temperature",
		},
		[]string{"target", "mac", "plugName"},
	)
	registry.MustRegister(metricOverTemp)

	metricWifiRssi := prometheus.NewGaugeVec(
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
	registry.MustRegister(metricWifiRssi)

	metricUpdateNeeded := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_update_needed",
			Help: "ShellyPlug status is update is needed",
		},
		[]string{"target", "mac", "plugName"},
	)
	registry.MustRegister(metricUpdateNeeded)

	metricCloudEnabled := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_cloud_enabled",
			Help: "ShellyPlug status if cloud is enabled",
		},
		[]string{"target", "mac", "plugName"},
	)
	registry.MustRegister(metricCloudEnabled)

	metricCloudConnected := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_cloud_connected",
			Help: "ShellyPlug status if device is connected to cloud",
		},
		[]string{"target", "mac", "plugName"},
	)
	registry.MustRegister(metricCloudConnected)

	metricPowerCurrent := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_power_current",
			Help: "ShellyPlug current power usage in watts",
		},
		[]string{"target", "mac", "plugName"},
	)
	registry.MustRegister(metricPowerCurrent)

	metricPowerTotal := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_power_total",
			Help: "ShellyPlug current power total in watts",
		},
		[]string{"target", "mac", "plugName"},
	)
	registry.MustRegister(metricPowerTotal)

	metricPowerLimit := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_power_limit",
			Help: "ShellyPlug configured power limit in watts",
		},
		[]string{"target", "mac", "plugName"},
	)
	registry.MustRegister(metricPowerLimit)

	metricSysUnixtime := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_system_unixtime",
			Help: "ShellyPlug system unixtime",
		},
		[]string{"target", "mac", "plugName"},
	)
	registry.MustRegister(metricSysUnixtime)

	metricSysUptime := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_system_uptime",
			Help: "ShellyPlug system uptime",
		},
		[]string{"target", "mac", "plugName"},
	)
	registry.MustRegister(metricSysUptime)

	metricSysMemTotal := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_system_memory_total",
			Help: "ShellyPlug system memory total",
		},
		[]string{"target", "mac", "plugName"},
	)
	registry.MustRegister(metricSysMemTotal)

	metricSysMemFree := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_system_memory_free",
			Help: "ShellyPlug system memory free",
		},
		[]string{"target", "mac", "plugName"},
	)
	registry.MustRegister(metricSysMemFree)

	metricSysFsSize := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_system_fs_size",
			Help: "ShellyPlug system filesystem size",
		},
		[]string{"target", "mac", "plugName"},
	)
	registry.MustRegister(metricSysFsSize)

	metricSysFsFree := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "shellyplug_system_fs_free",
			Help: "ShellyPlug system filesystem free",
		},
		[]string{"target", "mac", "plugName"},
	)
	registry.MustRegister(metricSysFsFree)

	client := resty.New()
	client.SetHeader("User-Agent", UserAgent+gitTag)

	if targetList, err := paramsGetListRequired(r.URL.Query(), "target"); err == nil {

		for _, target := range targetList {
			target = strings.TrimRight(target, "/")
			targetLogger := contextLogger.WithField("target", target)

			targetLabels := prometheus.Labels{
				"target":   target,
				"mac":      "",
				"plugName": "",
			}

			infoLabels := prometheus.Labels{
				"target":     target,
				"mac":        "",
				"hostname":   "",
				"plugName":   "",
				"plugType":   "",
				"plugSerial": "",
			}

			sp := shellyplug.New(target, client)
			if result, err := sp.GetSettings(); err == nil {
				targetLabels["plugName"] = result.Name
				targetLabels["mac"] = result.Device.Mac

				infoLabels["plugName"] = result.Name
				infoLabels["mac"] = result.Name
				infoLabels["hostname"] = result.Device.Hostname
				infoLabels["plugType"] = result.Device.Type

				metricPowerLimit.With(targetLabels).Set(result.MaxPower)
			}

			if result, err := sp.GetStatus(); err == nil {
				infoLabels["plugSerial"] = fmt.Sprintf("%d", result.Serial)
				metricInfo.With(infoLabels).Set(1)

				metricSysUnixtime.With(targetLabels).Set(float64(result.Unixtime))
				metricSysUptime.With(targetLabels).Set(float64(result.Uptime))
				metricSysMemTotal.With(targetLabels).Set(float64(result.RAMTotal))
				metricSysMemFree.With(targetLabels).Set(float64(result.RAMFree))
				metricSysFsSize.With(targetLabels).Set(float64(result.FsSize))
				metricSysFsFree.With(targetLabels).Set(float64(result.FsFree))

				metricTemp.With(targetLabels).Set(result.Temperature)
				metricOverTemp.With(targetLabels).Set(boolToFloat64(result.Overtemperature))

				wifiLabels := prometheus.Labels{
					"target":   target,
					"mac":      result.Mac,
					"plugName": targetLabels["plugName"],
					"ssid":     result.WifiSta.Ssid,
				}
				metricWifiRssi.With(wifiLabels).Set(float64(result.WifiSta.Rssi))

				metricUpdateNeeded.With(targetLabels).Set(boolToFloat64(result.HasUpdate))
				metricCloudEnabled.With(targetLabels).Set(boolToFloat64(result.Cloud.Enabled))
				metricCloudConnected.With(targetLabels).Set(boolToFloat64(result.Cloud.Connected))

				if len(result.Meters) >= 1 {
					powerUsage := result.Meters[0]
					metricPowerCurrent.With(targetLabels).Set(powerUsage.Power)
					metricPowerTotal.With(targetLabels).Set(powerUsage.Total)
				}

			} else {
				targetLogger.Errorf(`failed to fetch status: %v`, err)
			}

		}
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func paramsGetList(params url.Values, name string) (list []string, err error) {
	for _, v := range params[name] {
		list = append(list, strings.Split(v, ",")...)
	}
	return
}

func paramsGetListRequired(params url.Values, name string) (list []string, err error) {
	list, err = paramsGetList(params, name)

	if len(list) == 0 {
		err = fmt.Errorf("parameter \"%v\" is missing", name)
		return
	}

	return
}

func buildContextLoggerFromRequest(r *http.Request) *log.Entry {
	logFields := log.Fields{
		"requestPath": r.URL.Path,
	}

	return log.WithFields(logFields)
}

func getPrometheusTimeout(r *http.Request, defaultTimeout float64) (timeout float64, err error) {
	// If a timeout is configured via the Prometheus header, add it to the request.
	if v := r.Header.Get("X-Prometheus-Scrape-Timeout-Seconds"); v != "" {
		timeout, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return
		}
	}
	if timeout == 0 {
		timeout = defaultTimeout
	}

	return
}
