package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/webdevops/shelly-plug-exporter/shellyplug"
)

const (
	DefaultTimeout = 30
)

func newShellyProber(ctx context.Context, registry *prometheus.Registry, logger *zap.SugaredLogger) *shellyplug.ShellyPlug {
	sp := shellyplug.New(ctx, registry, logger)
	sp.SetUserAgent(UserAgent + gitTag)
	sp.SetTimeout(Opts.Shelly.Request.Timeout)
	if len(Opts.Shelly.Auth.Username) >= 1 {
		sp.SetHttpAuth(Opts.Shelly.Auth.Username, Opts.Shelly.Auth.Password)
	}

	return sp
}

func shellyProbeDiscovery(w http.ResponseWriter, r *http.Request) {
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

	sp := newShellyProber(ctx, registry, contextLogger)
	sp.UseDiscovery()
	sp.Run()

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func buildContextLoggerFromRequest(r *http.Request) *zap.SugaredLogger {
	return logger.With(zap.String("requestPath", r.URL.Path))
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
