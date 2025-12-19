package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/webdevops/go-common/log/slogger"

	"github.com/webdevops/shelly-plug-exporter/shellyplug"
)

const (
	DefaultTimeout = 30
)

func newShellyProber(ctx context.Context, registry *prometheus.Registry, logger *slogger.Logger) *shellyplug.ShellyPlug {
	sp := shellyplug.New(ctx, registry, logger)
	sp.SetUserAgent(UserAgent + gitTag)
	sp.SetTimeout(Opts.Shelly.Request.Timeout)
	if len(Opts.Shelly.Auth.Username) >= 1 {
		sp.SetHttpAuth(Opts.Shelly.Auth.Username, Opts.Shelly.Auth.Password)
	}

	return sp
}

func shellyProbeDiscoveryTargets(w http.ResponseWriter, r *http.Request) {
	registry := prometheus.NewRegistry()

	contextLogger := buildContextLoggerFromRequest(r)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30*float64(time.Second)))
	defer cancel()

	sp := newShellyProber(ctx, registry, contextLogger)
	sp.UseDiscovery()

	targets := sp.GetTargets()

	body, err := json.Marshal(targets)
	if err != nil {
		contextLogger.Error("failed to marshal object", slog.Any("error", err))
		http.Error(w, fmt.Sprintf("unable to marshal targets to json: %s", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(body)
	if err != nil {
		contextLogger.Error("failed to write", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
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
		contextLogger.Error("failed to detect prometheus timeout", slog.Any("error", err))
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

func buildContextLoggerFromRequest(r *http.Request) *slogger.Logger {
	return logger.With(slog.Group("request", slog.String("path", r.URL.Path)))
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
