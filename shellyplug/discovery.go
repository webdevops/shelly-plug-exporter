package shellyplug

import (
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/mdns"
	cache "github.com/patrickmn/go-cache"
)

const (
	ServiceDiscoveryCacheKey = "mdns:servicediscovery"
)

var (
	sdCache *cache.Cache
	sdLock  sync.Mutex
)

func EnableServiceDiscoveryCache(ttl time.Duration) {
	sdCache = cache.New(1*time.Minute, ttl)
}

func (sp *ShellyPlug) UseDiscovery(ttl time.Duration) {
	// try cache (without lock)
	if sdCache != nil {
		if val, ok := sdCache.Get(ServiceDiscoveryCacheKey); ok {
			sp.SetTargets(val.([]string))
			return
		}
	}

	sdLock.Lock()
	defer sdLock.Unlock()

	// try cache again, maybe last run set the cache again -> avoid another mdns sd run
	if sdCache != nil {
		if val, ok := sdCache.Get(ServiceDiscoveryCacheKey); ok {
			sp.SetTargets(val.([]string))
			return
		}
	}

	wg := sync.WaitGroup{}
	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, 4)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var targetList []string
		for entry := range entriesCh {
			if strings.HasPrefix(strings.ToLower(entry.Name), "shellyplug-") {
				sp.logger.Debugf(`found %v [%v] via mDNS servicediscovery`, entry.Name, entry.AddrV4.String())
				targetList = append(targetList, entry.AddrV4.String())
			}
		}

		if sdCache != nil {
			sdCache.SetDefault(ServiceDiscoveryCacheKey, targetList)
		}
		sp.SetTargets(targetList)
	}()

	// Start the lookup
	params := mdns.DefaultParams("_http._tcp")
	params.DisableIPv6 = true
	params.Timeout = ttl
	params.Entries = entriesCh
	err := mdns.Query(params)
	if err != nil {
		panic(err)
	}
	close(entriesCh)

	wg.Wait()
}
