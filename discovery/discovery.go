package discovery

import (
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/mdns"
	log "github.com/sirupsen/logrus"
)

const (
	TargetHealthDead = 0
	TargetHealthLow  = 1
	TargetHealthGood = 10

	TargetHealthy   = true
	TargetUnhealthy = false

	TargetTypeShellyPlug = "shellyplug"
	TargetTypeShellyPlus = "shellyplus"
	TargetTypeShellyPro  = "shellypro"
	TargetTypeShellyEm3  = "shellyem3"
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
	entriesCh := make(chan *mdns.ServiceEntry, 15)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var targetList []DiscoveryTarget
		for entry := range entriesCh {
			switch {
			case strings.HasPrefix(strings.ToLower(entry.Name), "shellyplug-"):
				targetList = append(targetList, createDiscoveryTarget(d, TargetTypeShellyPlug, entry))
			case strings.HasPrefix(strings.ToLower(entry.Name), "shellyplus"):
				targetList = append(targetList, createDiscoveryTarget(d, TargetTypeShellyPlus, entry))
			case strings.HasPrefix(strings.ToLower(entry.Name), "shellypro"):
				targetList = append(targetList, createDiscoveryTarget(d, TargetTypeShellyPro, entry))
			case strings.HasPrefix(strings.ToLower(entry.Name), "shellyem3"):
				targetList = append(targetList, createDiscoveryTarget(d, TargetTypeShellyEm3, entry))
			}
		}

		d.lock.Lock()
		defer d.lock.Unlock()

		// set all non-discovered targets to low health
		for target := range d.targetList {
			d.targetList[target].Health = TargetHealthLow
		}

		// set all discovered targets to good health
		for _, row := range targetList {
			target := row
			d.targetList[target.Address] = &target
			d.targetList[target.Address].Health = TargetHealthGood
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

func createDiscoveryTarget(d *serviceDiscovery, TargetType string, entry *mdns.ServiceEntry) DiscoveryTarget {
	d.logger.Debugf(`found %v [%v] via mDNS servicediscovery`, entry.Name, entry.AddrV4.String())
	return DiscoveryTarget{
		Hostname: entry.Name,
		Port:     entry.Port,
		Address:  entry.AddrV4.String(),
		Type:     TargetType,
	}
}

func (d *serviceDiscovery) MarkTarget(address string, healthy bool) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if target, exists := d.targetList[address]; exists {
		if healthy {
			d.targetList[address].Health = TargetHealthGood
		} else {
			d.targetList[address].Health = (target.Health - 1)
		}
	} else {
		if healthy {
			d.targetList[address].Health = TargetHealthGood
		}
	}

	d.cleanup()
}

func (d *serviceDiscovery) cleanup() {
	for address, target := range d.targetList {
		if target.Health <= TargetHealthDead {
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
