package discovery

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/mdns"
	"go.uber.org/zap"
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
)

type (
	serviceDiscovery struct {
		logger      *zap.SugaredLogger
		targetList  map[string]*DiscoveryTarget
		lock        sync.RWMutex
		staticHosts []DiscoveryTarget
	}
)

var (
	ServiceDiscovery *serviceDiscovery
)

func EnableDiscovery(logger *zap.SugaredLogger, refreshTime time.Duration, timeout time.Duration, shellyplugs []string, shellyplus []string, shellypro []string) {
	ServiceDiscovery = &serviceDiscovery{}
	ServiceDiscovery.logger = logger
	ServiceDiscovery.init(shellyplugs, shellyplus, shellypro)

	go func() {
		for {
			ServiceDiscovery.Run(timeout)
			time.Sleep(refreshTime)
		}
	}()
}

func (d *serviceDiscovery) init(shellyplugs []string, shellyplus []string, shellypro []string) {
	d.targetList = map[string]*DiscoveryTarget{}
	var staticHosts []DiscoveryTarget
	for _, entry := range shellyplugs {
		if entry != "" {
			staticHosts = append(staticHosts, discoveryTargetFromStatic(entry, TargetTypeShellyPlug))
		}
	}
	for _, entry := range shellyplus {
		if entry != "" {
			staticHosts = append(staticHosts, discoveryTargetFromStatic(entry, TargetTypeShellyPlus))
		}
	}
	for _, entry := range shellypro {
		if entry != "" {
			staticHosts = append(staticHosts, discoveryTargetFromStatic(entry, TargetTypeShellyPro))
		}
	}
	d.staticHosts = staticHosts
}

func (d *serviceDiscovery) Run(timeout time.Duration) {
	wg := sync.WaitGroup{}
	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, 15)

	wg.Add(1)
	go func() {
		defer wg.Done()
		var targetList []DiscoveryTarget
		targetList = append(targetList, d.staticHosts...)
		for entry := range entriesCh {
			switch {
			case strings.HasPrefix(strings.ToLower(entry.Name), "shellyplug-"):
				d.logger.Debugf(`found %v [%v] via mDNS servicediscovery`, entry.Name, entry.AddrV4.String())
				targetList = append(targetList, DiscoveryTarget{
					Hostname: entry.Name,
					Port:     entry.Port,
					Address:  entry.AddrV4.String(),
					Type:     TargetTypeShellyPlug,
					Static:   false,
				})
			case strings.HasPrefix(strings.ToLower(entry.Name), "shellyplus"):
				d.logger.Debugf(`found %v [%v] via mDNS servicediscovery`, entry.Name, entry.AddrV4.String())
				targetList = append(targetList, DiscoveryTarget{
					Hostname: entry.Name,
					Port:     entry.Port,
					Address:  entry.AddrV4.String(),
					Type:     TargetTypeShellyPlus,
					Static:   false,
				})
			case strings.HasPrefix(strings.ToLower(entry.Name), "shellypro"):
				d.logger.Debugf(`found %v [%v] via mDNS servicediscovery`, entry.Name, entry.AddrV4.String())
				targetList = append(targetList, DiscoveryTarget{
					Hostname: entry.Name,
					Port:     entry.Port,
					Address:  entry.AddrV4.String(),
					Type:     TargetTypeShellyPro,
					Static:   false,
				})
			}
		}
		d.logger.Debugf(`Total targets: "%v"`, targetList)

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

func (d *serviceDiscovery) MarkTarget(address string, healthy bool) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if target, exists := d.targetList[address]; exists {
		if target.Static {
			// Assume Static targets are always healthy
			return
		}
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

func discoveryTargetFromStatic(entry string, typ string) DiscoveryTarget {
	parts := strings.Split(entry, ":")
	var name string
	var port int
	if len(parts) == 2 {
		name = parts[0]
		var err error
		port, err = strconv.Atoi(parts[1])
		if err != nil {
			panic(err)
		}
	} else {
		name = parts[0]
		port = 80
	}

	return DiscoveryTarget{
		Hostname: name,
		Port:     port,
		Address:  name,
		Type:     typ,
		Static:   true,
		Health:   TargetHealthGood,
	}
}
