package shellyplug

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type (
	ShellyPlug struct {
		ctx      context.Context
		client   *resty.Client
		logger   *log.Entry
		registry *prometheus.Registry

		targets struct {
			list []string
			lock sync.RWMutex
		}

		prometheus struct {
			info         *prometheus.GaugeVec
			temp         *prometheus.GaugeVec
			overTemp     *prometheus.GaugeVec
			wifiRssi     *prometheus.GaugeVec
			updateNeeded *prometheus.GaugeVec

			cloudEnabled   *prometheus.GaugeVec
			cloudConnected *prometheus.GaugeVec

			switchOn        *prometheus.GaugeVec
			switchOverpower *prometheus.GaugeVec
			switchTimer     *prometheus.GaugeVec

			powerCurrent *prometheus.GaugeVec
			powerTotal   *prometheus.GaugeVec
			powerLimit   *prometheus.GaugeVec

			sysUnixtime *prometheus.GaugeVec
			sysUptime   *prometheus.GaugeVec
			sysMemTotal *prometheus.GaugeVec
			sysMemFree  *prometheus.GaugeVec
			sysFsSize   *prometheus.GaugeVec
			sysFsFree   *prometheus.GaugeVec
		}
	}
)

func New(ctx context.Context, registry *prometheus.Registry, logger *log.Entry) *ShellyPlug {
	sp := ShellyPlug{}
	sp.ctx = ctx
	sp.registry = registry
	sp.logger = logger
	sp.initResty()
	sp.initMetrics()
	return &sp
}

func (sp *ShellyPlug) initResty() {
	sp.client = resty.New()

	sp.client.OnAfterResponse(func(client *resty.Client, response *resty.Response) error {
		switch statusCode := response.StatusCode(); statusCode {
		case 401:
			return errors.New(`shelly plug requires authentication and/or credentials are invalid`)
		case 200:
			// all ok, proceed
			return nil
		default:
			return fmt.Errorf(`expected http status 200, got %v`, response.StatusCode())
		}
	})
}

func (sp *ShellyPlug) SetUserAgent(val string) {
	sp.client.SetHeader("User-Agent", val)
}

func (sp *ShellyPlug) SetTimeout(timeout time.Duration) {
	sp.client.SetTimeout(timeout)
}

func (sp *ShellyPlug) SetHttpAuth(username, password string) {
	sp.client.SetDisableWarn(true)
	sp.client.SetBasicAuth(username, password)
}

func (sp *ShellyPlug) SetTargets(targets []string) {
	sp.targets.lock.Lock()
	defer sp.targets.lock.Unlock()
	sp.targets.list = targets
}

func (sp *ShellyPlug) GetTargets() []string {
	sp.targets.lock.RLock()
	defer sp.targets.lock.RUnlock()

	return sp.targets.list
}

func (sp *ShellyPlug) Run() {
	wg := sync.WaitGroup{}

	for _, target := range sp.GetTargets() {
		wg.Add(1)
		go func(target string) {
			defer wg.Done()
			sp.collectFromTarget(target)
		}(target)
	}
	wg.Wait()
}

func (sp *ShellyPlug) collectFromTarget(target string) {
	targetLogger := sp.logger.WithField("target", target)

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

	if result, err := sp.targetGetSettings(target); err == nil {
		if discovery != nil {
			discovery.MarkTarget(target, DiscoveryTargetHealthy)
		}

		targetLabels["plugName"] = result.Name
		targetLabels["mac"] = result.Device.Mac

		infoLabels["plugName"] = result.Name
		infoLabels["mac"] = result.Device.Mac
		infoLabels["hostname"] = result.Device.Hostname
		infoLabels["plugType"] = result.Device.Type

		sp.prometheus.powerLimit.With(targetLabels).Set(result.MaxPower)
	} else {
		targetLogger.Errorf(`failed to fetch settings: %v`, err)
		if discovery != nil {
			discovery.MarkTarget(target, DiscoveryTargetUnhealthy)
		}
	}

	if result, err := sp.targetGetStatus(target); err == nil {
		infoLabels["plugSerial"] = fmt.Sprintf("%d", result.Serial)
		sp.prometheus.info.With(infoLabels).Set(1)

		sp.prometheus.sysUnixtime.With(targetLabels).Set(float64(result.Unixtime))
		sp.prometheus.sysUptime.With(targetLabels).Set(float64(result.Uptime))
		sp.prometheus.sysMemTotal.With(targetLabels).Set(float64(result.RAMTotal))
		sp.prometheus.sysMemFree.With(targetLabels).Set(float64(result.RAMFree))
		sp.prometheus.sysFsSize.With(targetLabels).Set(float64(result.FsSize))
		sp.prometheus.sysFsFree.With(targetLabels).Set(float64(result.FsFree))

		sp.prometheus.temp.With(targetLabels).Set(result.Temperature)
		sp.prometheus.overTemp.With(targetLabels).Set(boolToFloat64(result.Overtemperature))

		wifiLabels := prometheus.Labels{
			"target":   target,
			"mac":      targetLabels["mac"],
			"plugName": targetLabels["plugName"],
			"ssid":     result.WifiSta.Ssid,
		}
		sp.prometheus.wifiRssi.With(wifiLabels).Set(float64(result.WifiSta.Rssi))

		sp.prometheus.updateNeeded.With(targetLabels).Set(boolToFloat64(result.HasUpdate))
		sp.prometheus.cloudEnabled.With(targetLabels).Set(boolToFloat64(result.Cloud.Enabled))
		sp.prometheus.cloudConnected.With(targetLabels).Set(boolToFloat64(result.Cloud.Connected))

		if len(result.Meters) >= 1 {
			powerUsage := result.Meters[0]
			sp.prometheus.powerCurrent.With(targetLabels).Set(powerUsage.Power)
			// total is provided as watt/minutes, we want watt/hours
			sp.prometheus.powerTotal.With(targetLabels).Set(powerUsage.Total / 60)
		}

		if len(result.Relays) >= 1 {
			relay := result.Relays[0]

			switchLabels := prometheus.Labels{
				"target":       targetLabels["target"],
				"mac":          targetLabels["mac"],
				"plugName":     targetLabels["plugName"],
				"switchSource": relay.Source,
			}

			sp.prometheus.switchOn.With(switchLabels).Set(boolToFloat64(relay.Ison))
			sp.prometheus.switchOverpower.With(targetLabels).Set(boolToFloat64(relay.Overpower))
			sp.prometheus.switchTimer.With(targetLabels).Set(boolToFloat64(relay.HasTimer))
		}

	} else {
		targetLogger.Errorf(`failed to fetch status: %v`, err)
		if discovery != nil {
			discovery.MarkTarget(target, DiscoveryTargetUnhealthy)
		}
	}
}

func (sp *ShellyPlug) targetGetSettings(target string) (ResultSettings, error) {
	url := fmt.Sprintf("http://%v/settings", target)

	result := ResultSettings{}

	r := sp.client.R().SetContext(sp.ctx).SetResult(&result).ForceContentType("application/json")
	_, err := r.Get(url)
	return result, err
}

func (sp *ShellyPlug) targetGetStatus(target string) (ResultStatus, error) {
	url := fmt.Sprintf("http://%v/status", target)

	result := ResultStatus{}

	r := sp.client.R().SetContext(sp.ctx).SetResult(&result).ForceContentType("application/json")
	_, err := r.Get(url)
	return result, err
}
