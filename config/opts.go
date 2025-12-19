package config

import (
	"encoding/json"
	"time"
)

type (
	Opts struct {
		// logger
		Logger struct {
			Level  string `long:"log.level"    env:"LOG_LEVEL"   description:"Log level" choice:"trace" choice:"debug" choice:"info" choice:"warning" choice:"error" default:"info"`                          // nolint:staticcheck // multiple choices are ok
			Format string `long:"log.format"   env:"LOG_FORMAT"  description:"Log format" choice:"logfmt" choice:"json" default:"logfmt"`                                                                     // nolint:staticcheck // multiple choices are ok
			Source string `long:"log.source"   env:"LOG_SOURCE"  description:"Show source for every log message (useful for debugging and bug reports)" choice:"" choice:"short" choice:"file" choice:"full"` // nolint:staticcheck // multiple choices are ok
			Color  string `long:"log.color"    env:"LOG_COLOR"   description:"Enable color for logs" choice:"" choice:"auto" choice:"yes" choice:"no"`                                                        // nolint:staticcheck // multiple choices are ok
			Time   bool   `long:"log.time"     env:"LOG_TIME"    description:"Show log time"`
		}

		Shelly struct {
			Request struct {
				Timeout time.Duration `long:"shelly.request.timeout"  env:"SHELLY_REQUEST_TIMEOUT"  description:"Request timeout" default:"5s"`
			}

			Auth struct {
				Username string `long:"shelly.auth.username"  env:"SHELLY_AUTH_USERNAME"  description:"Username for shelly plug login"`
				Password string `long:"shelly.auth.password"  env:"SHELLY_AUTH_PASSWORD"  description:"Password for shelly plug login" json:"-"`
			}

			Host struct {
				ShellyPlug []string `long:"shelly.host.shellyplug"  env:"SHELLY_HOST_SHELLYPLUGS" env-delim:","  description:"shellyplug device IP or hostname to scrape. Pass multiple times for multiple hosts" default:""`
				ShellyPlus []string `long:"shelly.host.shellyplus"  env:"SHELLY_HOST_SHELLYPLUSES" env-delim:","  description:"shellyplus device IP or hostname to scrape. Pass multiple times for multiple hosts" default:""`
				ShellyPro  []string `long:"shelly.host.shellypro"  env:"SHELLY_HOST_SHELLYPROS" env-delim:","  description:"shellypro device IP or hostname to scrape. Pass multiple times for multiple hosts" default:""`
			}

			ServiceDiscovery struct {
				Timeout time.Duration `long:"shelly.servicediscovery.timeout"  env:"SHELLY_SERVICEDISCOVERY_TIMEOUT"  description:"mDNS discovery response timeout" default:"15s"`
				Refresh time.Duration `long:"shelly.servicediscovery.refresh"  env:"SHELLY_SERVICEDISCOVERY_REFRESH"  description:"mDNS discovery refresh time" default:"15m"`
			}
		}

		// general options
		Server struct {
			// general options
			Bind         string        `long:"server.bind"              env:"SERVER_BIND"           description:"Server address"        default:":8080"`
			ReadTimeout  time.Duration `long:"server.timeout.read"      env:"SERVER_TIMEOUT_READ"   description:"Server read timeout"   default:"5s"`
			WriteTimeout time.Duration `long:"server.timeout.write"     env:"SERVER_TIMEOUT_WRITE"  description:"Server write timeout"  default:"10s"`
		}
	}
)

func (o *Opts) GetJson() []byte {
	jsonBytes, err := json.Marshal(o)
	if err != nil {
		panic(err)
	}
	return jsonBytes
}
