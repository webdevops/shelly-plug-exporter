package shellyplug

import (
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/mdns"
	log "github.com/sirupsen/logrus"
)

const (
	DiscoveryTargetHealthDead = 0
	DiscoveryTargetHealthLow  = 1
	DiscoveryTargetHealthGood = 10

	DiscoveryTargetHealthy   = true
	DiscoveryTargetUnhealthy = false
)

type (
	Discovery struct {
		logger     *log.Entry
		targetList map[string]int
		lock       sync.RWMutex
	}
)

var (
	discovery *Discovery
)

func EnableDiscovery(logger *log.Entry, refreshTime time.Duration, timeout time.Duration) {
	discovery = &Discovery{}
	discovery.logger = logger
	discovery.init()

	go func() {
		for {
			discovery.Run(timeout)
			time.Sleep(refreshTime)
		}
	}()
}

func (d *Discovery) init() {
	d.targetList = map[string]int{}
}

func (d *Discovery) Run(timeout time.Duration) {
	wg := sync.WaitGroup{}
	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, 4)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var targetList []string
		for entry := range entriesCh {
			if strings.HasPrefix(strings.ToLower(entry.Name), "shellyplug-") {
				d.logger.Debugf(`found %v [%v] via mDNS servicediscovery`, entry.Name, entry.AddrV4.String())
				targetList = append(targetList, entry.AddrV4.String())
			}
		}

		d.lock.Lock()
		defer d.lock.Unlock()

		// set all non discovered targets to low health
		for target := range d.targetList {
			d.targetList[target] = DiscoveryTargetHealthLow
		}

		// set all discovered targets to good health
		for _, target := range targetList {
			d.targetList[target] = DiscoveryTargetHealthGood
		}
		d.cleanup()
	}()

	// Start the lookup
	params := mdns.DefaultParams("_http._tcp")
	params.DisableIPv6 = true
	params.Timeout = timeout
	params.Entries = entriesCh
	err := mdns.Query(params)
	if err != nil {
		panic(err)
	}
	close(entriesCh)

	wg.Wait()
}

func (d *Discovery) MarkTarget(target string, healthy bool) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if healthCount, exists := d.targetList[target]; exists {
		if healthy {
			d.targetList[target] = DiscoveryTargetHealthGood
		} else {
			d.targetList[target] = (healthCount - 1)
		}
	} else {
		if healthy {
			d.targetList[target] = DiscoveryTargetHealthGood
		}
	}

	d.cleanup()
}

func (d *Discovery) cleanup() {
	for target, health := range d.targetList {
		if health <= DiscoveryTargetHealthDead {
			d.logger.Debugf(`disabling unhealthy target "%v"`, target)
			delete(d.targetList, target)
		}
	}
}

func (d *Discovery) GetTargetList() []string {
	d.lock.RLock()
	defer d.lock.RUnlock()

	targetList := []string{}
	for target := range d.targetList {
		targetList = append(targetList, target)
	}

	return targetList
}

func (sp *ShellyPlug) UseDiscovery() {
	if discovery == nil {
		panic(`servicediscovery not enabled`)
	}

	sp.SetTargets(discovery.GetTargetList())
}
