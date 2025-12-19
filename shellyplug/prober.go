package shellyplug

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/webdevops/go-common/log/slogger"

	"github.com/webdevops/shelly-plug-exporter/discovery"
)

type (
	ShellyPlug struct {
		ctx      context.Context
		logger   *slogger.Logger
		registry *prometheus.Registry

		auth struct {
			username string
			password string
		}

		resty struct {
			timeout   time.Duration
			userAgent string
		}

		targets struct {
			list []discovery.DiscoveryTarget
			lock sync.RWMutex
		}

		prometheus shellyPlugMetrics
	}
)

func New(ctx context.Context, registry *prometheus.Registry, logger *slogger.Logger) *ShellyPlug {
	sp := ShellyPlug{}
	sp.ctx = ctx
	sp.registry = registry
	sp.logger = logger
	sp.initMetrics()

	if globalCache == nil {
		globalCache = cache.New(15*time.Minute, 1*time.Minute)
	}

	if restyCache == nil {
		restyCache = cache.New(1*time.Hour, 1*time.Minute)
	}

	return &sp
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

	for _, row := range sp.GetTargets() {
		target := row
		wg.Add(1)
		go func(target discovery.DiscoveryTarget) {
			defer wg.Done()
			sp.collectFromTarget(target)
		}(target)
	}
	wg.Wait()
}

func (sp *ShellyPlug) collectFromTarget(target discovery.DiscoveryTarget) {
	targetLogger := sp.logger.With(
		slog.Group(
			"target",
			slog.String("name", target.Name()),
			slog.String("address", target.Address),
		),
	)

	targetLogger.Debug("probing shelly device")

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
			targetLogger.Debug(`detected shelly device generation`, slog.Int("gen", *result.Gen))
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
		targetLogger.Error(`failed to fetch settings`, slog.Any("error", err))
		if discovery.ServiceDiscovery != nil {
			discovery.ServiceDiscovery.MarkTarget(target.Address, discovery.TargetUnhealthy)
		}
	}

	targetLogger = targetLogger.With(slog.Int("gen", shellyGeneration))
	switch shellyGeneration {
	case 1:
		sp.collectFromTargetGen1(target, targetLogger, infoLabels, targetLabels)
	case 2:
		sp.collectFromTargetGen2(target, targetLogger, infoLabels, targetLabels)
	default:
		targetLogger.Warn("unsupported Shelly generation", slog.Int("gen", shellyGeneration))
	}
}
