package config

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

type (
	Opts struct {
		// logger
		Logger struct {
			Debug   bool `           long:"debug"        env:"DEBUG"    description:"debug mode"`
			Verbose bool `short:"v"  long:"verbose"      env:"VERBOSE"  description:"verbose mode"`
			LogJson bool `           long:"log.json"     env:"LOG_JSON" description:"Switch log output to json format"`
		}

		Shelly struct {
			Auth struct {
				Username string `long:"shelly.auth.username"  env:"SHELLY_AUTH_USERNAME"  description:"Username for shelly plug login"`
				Password string `long:"shelly.auth.password"  env:"SHELLY_AUTH_PASSWORD"  description:"Password for shelly plug login"`
			}

			ServiceDiscovery struct {
				Timeout time.Duration `long:"shelly.servicediscovery.timeout"  env:"SHELLY_SERVICEDISCOVERY_TIMEOUT"  description:"mDNS discovery response timeout" default:"5s"`

				Cache struct {
					Ttl time.Duration `long:"shelly.servicediscovery.cache.ttl"  env:"SHELLY_SERVICEDISCOVERY_CACHE_TTL"  description:"mDNS discovery cache ttl" default:"15m"`
				}
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
		log.Panic(err)
	}
	return jsonBytes
}
