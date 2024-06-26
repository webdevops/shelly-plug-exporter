package shellyplug

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/icholy/digest"
	cache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/webdevops/shelly-plug-exporter/discovery"
)

type (
	ShellyPlug struct {
		ctx      context.Context
		client   *resty.Client
		logger   *zap.SugaredLogger
		registry *prometheus.Registry

		targets struct {
			list []discovery.DiscoveryTarget
			lock sync.RWMutex
		}

		prometheus shellyPlugMetrics
	}
)

func New(ctx context.Context, registry *prometheus.Registry, logger *zap.SugaredLogger) *ShellyPlug {
	sp := ShellyPlug{}
	sp.ctx = ctx
	sp.registry = registry
	sp.logger = logger
	sp.initResty()
	sp.initMetrics()

	if globalCache == nil {
		globalCache = cache.New(15*time.Minute, 1*time.Minute)
	}

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
	sp.client.SetBasicAuth(username, password).SetTransport(
		&digest.Transport{
			Username: username,
			Password: password,
		},
	)
}

func (sp *ShellyPlug) SetTargets(targets []discovery.DiscoveryTarget) {
	sp.targets.lock.Lock()
	defer sp.targets.lock.Unlock()
	sp.targets.list = targets
}

func (sp *ShellyPlug) GetTargets() []discovery.DiscoveryTarget {
	sp.targets.lock.RLock()
	defer sp.targets.lock.RUnlock()

	return sp.targets.list
}

func (sp *ShellyPlug) Run() {
	wg := sync.WaitGroup{}

	for _, target := range sp.GetTargets() {
		wg.Add(1)
		go func(target discovery.DiscoveryTarget) {
			defer wg.Done()
			sp.collectFromTarget(target)
		}(target)
	}
	wg.Wait()
}

func (sp *ShellyPlug) collectFromTarget(target discovery.DiscoveryTarget) {
	targetLogger := sp.logger.With(zap.String("target", target.Address))

	targetLogger.Debugf("probing shelly %v", target.Name())

	targetLabels := prometheus.Labels{
		"target":   target.Address,
		"mac":      "",
		"plugName": "",
	}

	infoLabels := prometheus.Labels{
		"target":         target.Address,
		"mac":            "",
		"hostname":       "",
		"plugName":       "",
		"plugModel":      "",
		"plugApp":        "",
		"plugGeneration": "",
	}

	shellyGeneration := 0
	if result, err := sp.targetGetShellyInfo(target); err == nil {
		if result.Gen != nil {
			targetLogger.Debugf(`detected shelly device generation %v`, *result.Gen)
			shellyGeneration = *result.Gen
		} else {
			shellyGeneration = 1
		}

		targetLabels["plugName"] = result.Name
		targetLabels["mac"] = result.Mac

		infoLabels["plugName"] = result.Name
		infoLabels["mac"] = result.Mac
		infoLabels["hostname"] = target.Hostname
		infoLabels["plugModel"] = result.Model
		infoLabels["plugApp"] = result.App
		infoLabels["plugGeneration"] = strconv.Itoa(shellyGeneration)

	} else {
		targetLogger.Errorf(`failed to fetch settings: %v`, err)
		if discovery.ServiceDiscovery != nil {
			discovery.ServiceDiscovery.MarkTarget(target.Address, discovery.TargetUnhealthy)
		}
	}

	switch shellyGeneration {
	case 1:
		sp.collectFromTargetGen1(target, targetLogger, infoLabels, targetLabels)
	case 2:
		sp.collectFromTargetGen2(target, targetLogger, infoLabels, targetLabels)
	default:
		targetLogger.Warnf("unsupported Shelly generation %v", shellyGeneration)
	}
}
