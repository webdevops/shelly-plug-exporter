package discovery

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
	serviceDiscovery struct {
		logger     *log.Entry
		targetList map[string]*DiscoveryTarget
		lock       sync.RWMutex
	}
)

var (
	ServiceDiscovery *serviceDiscovery
)

func EnableDiscovery(logger *log.Entry, refreshTime time.Duration, timeout time.Duration) {
	ServiceDiscovery = &serviceDiscovery{}
	ServiceDiscovery.logger = logger
	ServiceDiscovery.init()

	go func() {
		for {
			ServiceDiscovery.Run(timeout)
			time.Sleep(refreshTime)
		}
	}()
}

func (d *serviceDiscovery) init() {
	d.targetList = map[string]*DiscoveryTarget{}
}

func (d *serviceDiscovery) Run(timeout time.Duration) {
	wg := sync.WaitGroup{}
	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, 4)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var targetList []DiscoveryTarget
		for entry := range entriesCh {
			switch {
			case strings.HasPrefix(strings.ToLower(entry.Name), "shellyplug-"):
				d.logger.Debugf(`found %v [%v] via mDNS servicediscovery`, entry.Name, entry.AddrV4.String())
				targetList = append(targetList, DiscoveryTarget{
					Hostname: entry.Name,
					Port:     entry.Port,
					Address:  entry.AddrV4.String(),
				})
			case strings.HasPrefix(strings.ToLower(entry.Name), "shellyplus"):
				d.logger.Debugf(`found %v [%v] via mDNS servicediscovery`, entry.Name, entry.AddrV4.String())
				targetList = append(targetList, DiscoveryTarget{
					Hostname: entry.Name,
					Port:     entry.Port,
					Address:  entry.AddrV4.String(),
				})
			}
		}

		d.lock.Lock()
		defer d.lock.Unlock()

		// set all non-discovered targets to low health
		for target := range d.targetList {
			d.targetList[target].Health = DiscoveryTargetHealthLow
		}

		// set all discovered targets to good health
		for _, row := range targetList {
			target := row
			d.targetList[target.Address] = &target
			d.targetList[target.Address].Health = DiscoveryTargetHealthGood
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

func (d *serviceDiscovery) MarkTarget(address string, healthy bool) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if target, exists := d.targetList[address]; exists {
		if healthy {
			d.targetList[address].Health = DiscoveryTargetHealthGood
		} else {
			d.targetList[address].Health = (target.Health - 1)
		}
	} else {
		if healthy {
			d.targetList[address].Health = DiscoveryTargetHealthGood
		}
	}

	d.cleanup()
}

func (d *serviceDiscovery) cleanup() {
	for address, target := range d.targetList {
		if target.Health <= DiscoveryTargetHealthDead {
			d.logger.Debugf(`disabling unhealthy target "%v"`, target.Name())
			delete(d.targetList, address)
		}
	}
}

func (d *serviceDiscovery) GetTargetList() []DiscoveryTarget {
	d.lock.RLock()
	defer d.lock.RUnlock()

	targetList := []DiscoveryTarget{}

	for _, row := range d.targetList {
		target := *row
		targetList = append(targetList, target)
	}

	return targetList
}
