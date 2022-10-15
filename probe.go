package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"

	"github.com/webdevops/shelly-plug-exporter/shellyplug"
)

const (
	DefaultTimeout = 30
)

func newShellyuProber(ctx context.Context, registry *prometheus.Registry, logger *log.Entry) *shellyplug.ShellyPlug {
	sp := shellyplug.New(ctx, registry, logger)
	sp.SetUserAgent(UserAgent + gitTag)
	sp.SetTimeout(opts.Shelly.Request.Timeout)
	if len(opts.Shelly.Auth.Username) >= 1 {
		sp.SetHttpAuth(opts.Shelly.Auth.Username, opts.Shelly.Auth.Password)
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

	sp := newShellyuProber(ctx, registry, contextLogger)
	sp.UseDiscovery()
	sp.Run()

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func shellyProbeTargets(w http.ResponseWriter, r *http.Request) {
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

	sp := newShellyuProber(ctx, registry, contextLogger)

	if targetList, err := paramsGetListRequired(r.URL.Query(), "target"); err == nil {
		sp.SetTargets(targetList)
		sp.Run()
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
