package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/webdevops/shelly-plug-exporter/config"
	"github.com/webdevops/shelly-plug-exporter/discovery"
)

const (
	Author    = "webdevops.io"
	UserAgent = "shelly-plug-exporter/"
)

var (
	argparser *flags.Parser
	Opts      config.Opts

	// Git version information
	gitCommit = "<unknown>"
	gitTag    = "<unknown>"
)

func main() {
	initArgparser()
	initLogger()

	logger.Infof("starting shellyplug-plug-exporter v%s (%s; %s; by %v)", gitTag, gitCommit, runtime.Version(), Author)
	logger.Info(string(Opts.GetJson()))

	logger.Infof("starting http server on %s", Opts.Server.Bind)
	startHttpServer()
}

// init argparser and parse/validate arguments
func initArgparser() {
	argparser = flags.NewParser(&Opts, flags.Default)
	_, err := argparser.Parse()

	// check if there is an parse error
	if err != nil {
		var flagsErr *flags.Error
		if ok := errors.As(err, &flagsErr); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Println()
			argparser.WriteHelp(os.Stdout)
			os.Exit(1)
		}
	}
}

// start and handle prometheus handler
func startHttpServer() {
	mux := http.NewServeMux()

	// healthz
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "Ok"); err != nil {
			logger.Error(err)
		}
	})

	// readyz
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, "Ok"); err != nil {
			logger.Error(err)
		}
	})

	discovery.EnableDiscovery(
		logger.With(zap.String("module", "discovery")),
		Opts.Shelly.ServiceDiscovery.Refresh,
		Opts.Shelly.ServiceDiscovery.Timeout,
		Opts.Shelly.Host.ShellyPlug,
		Opts.Shelly.Host.ShellyPlus,
		Opts.Shelly.Host.ShellyPro,
	)

	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/probe", shellyProbeDiscovery)

	srv := &http.Server{
		Addr:         Opts.Server.Bind,
		Handler:      mux,
		ReadTimeout:  Opts.Server.ReadTimeout,
		WriteTimeout: Opts.Server.WriteTimeout,
	}
	logger.Fatal(srv.ListenAndServe())
}
