package shellyplug

import (
	"context"
	"errors"
	"fmt"
	"time"

	resty "github.com/go-resty/resty/v2"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"

	"github.com/webdevops/shelly-plug-exporter/discovery"
)

var (
	restyCache *cache.Cache
)

func (sp *ShellyPlug) SetUserAgent(val string) {
	sp.resty.userAgent = val
}

func (sp *ShellyPlug) SetTimeout(timeout time.Duration) {
	sp.resty.timeout = timeout
}

func (sp *ShellyPlug) SetHttpAuth(username, password string) {
	sp.auth.username = username
	sp.auth.password = password
}

func (sp *ShellyPlug) restyClient(ctx context.Context, target discovery.DiscoveryTarget) (client *resty.Client) {
	cacheKey := target.Address
	if val, ok := restyCache.Get(cacheKey); ok {
		if client, ok := val.(*resty.Client); ok {
			return client
		}
	}

	restyLogger := sp.logger.With(
		zap.String("target", target.Address),
		zap.String("type", target.Type),
	)

	client = resty.New()
	client.SetBaseURL(target.BaseUrl())
	client.SetLogger(restyLogger)
	if sp.resty.timeout.Seconds() > 0 {
		client.SetTimeout(sp.resty.timeout)
	}

	if sp.resty.userAgent != "" {
		client.SetHeader("User-Agent", sp.resty.userAgent)
	}
	// client.SetRetryCount(2).
	// 	SetRetryWaitTime(1 * time.Second).
	// 	SetRetryMaxWaitTime(5 * time.Second)

	if sp.auth.username != "" {
		client.SetDisableWarn(true)
		client.SetBasicAuth(sp.auth.username, sp.auth.password)
		client.SetDigestAuth(sp.auth.username, sp.auth.password)
	}

	// client.AddRequestMiddleware(func(c *resty.Client, req *resty.Request) error {
	// 	c.Logger().Debugf(`send request: %s %s`, req.Method, req.URL)
	// 	return nil
	// })
	//
	// client.AddResponseMiddleware(func(c *resty.Client, res *resty.Response) error {
	// 	c.Logger().Debugf(`got response: %s %s with status %v`, res.Request.Method, res.Request.RawRequest.URL.String(), res.StatusCode())
	// 	return nil
	// })

	client.OnAfterResponse(func(c *resty.Client, res *resty.Response) error {
		switch res.StatusCode() {
		case 401:
			return errors.New(`shelly plug requires authentication and/or credentials are invalid`)
		case 200:
			// all ok, proceed
			return nil
		default:
			return fmt.Errorf(`expected http status 200, got %v`, res.StatusCode())
		}
	})

	restyCache.SetDefault(cacheKey, client)

	return
}
