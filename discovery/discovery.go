package discovery

import (
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/mdns"
	"github.com/webdevops/go-common/log/slogger"
)

const (
	TargetHealthDead = 0
	TargetHealthLow  = 2
	TargetHealthGood = 10

	TargetHealthy   = true
	TargetUnhealthy = false

	TargetTypeShellyPlug = "shellyplug"
	TargetTypeShellyPlus = "shellyplus"
	TargetTypeShellyPro  = "shellypro"
)

type (
	serviceDiscovery struct {
		logger      *slogger.Logger
		targetList  map[string]*DiscoveryTarget
		lock        sync.RWMutex
		staticHosts []DiscoveryTarget
	}

	serviceDiscoveryTarget struct {
		mdns.ServiceEntry

		Name    string
		Host    string
		Address string
		Port    int

		InfoFields map[string]string
		Generation string
		Version    string
	}
)

var (
	ServiceDiscovery *serviceDiscovery
)

func EnableDiscovery(logger *slogger.Logger, refreshTime time.Duration, timeout time.Duration, shellyplugs []string, shellyplus []string, shellypro []string) {
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
	var targetList []DiscoveryTarget
	targetList = append(targetList, d.staticHosts...)

	wg := sync.WaitGroup{}

	targetChannel := make(chan *DiscoveryTarget, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		for target := range targetChannel {
			targetList = append(targetList, *target)
		}
	}()

	// mDNS discovery via _http._tcp.
	d.discover("_http._tcp", timeout, func(logger *slogger.Logger, target *serviceDiscoveryTarget) *DiscoveryTarget {
		switch {
		case strings.HasPrefix(target.Name, "shellyplug-"):
			logger.Debug(`found target via mDNS servicediscovery`)
			return &DiscoveryTarget{
				Hostname:   target.Name,
				Port:       target.Port,
				Address:    target.Address,
				Type:       TargetTypeShellyPlug,
				Generation: target.Generation,
				Static:     false,
			}
		case strings.HasPrefix(target.Name, "shellyplus"):
			logger.Debug(`found target via mDNS servicediscovery`)
			return &DiscoveryTarget{
				Hostname:   target.Name,
				Port:       target.Port,
				Address:    target.Address,
				Type:       TargetTypeShellyPlus,
				Generation: target.Generation,
				Static:     false,
			}
		case strings.HasPrefix(target.Name, "shellypro"):
			logger.Debug(`found target via mDNS servicediscovery`)
			return &DiscoveryTarget{
				Hostname:   target.Name,
				Port:       target.Port,
				Address:    target.Address,
				Type:       TargetTypeShellyPro,
				Generation: target.Generation,
				Static:     false,
			}
		}

		if gen, ok := target.InfoFields["gen"]; ok {
			switch strings.ToLower(gen) {
			case "2":
				return &DiscoveryTarget{
					Hostname:   target.Name,
					Port:       target.Port,
					Address:    target.Address,
					Type:       TargetTypeShellyPro,
					Generation: target.Generation,
					Static:     false,
				}
			}
		}

		return nil
	}, targetChannel)

	// mDNS discovery via _shelly._tcp
	d.discover("_shelly._tcp", timeout, func(logger *slogger.Logger, target *serviceDiscoveryTarget) *DiscoveryTarget {
		if gen, ok := target.InfoFields["gen"]; ok {
			switch strings.ToLower(gen) {
			case "2":
				return &DiscoveryTarget{
					Hostname:   target.Name,
					Port:       target.Port,
					Address:    target.Address,
					Type:       TargetTypeShellyPro,
					Generation: target.Generation,
					Static:     false,
				}
			}
		}

		return nil
	}, targetChannel)

	close(targetChannel)
	wg.Wait()

	d.lock.Lock()
	defer d.lock.Unlock()

	// reduce all non-discovered targets health
	for target := range d.targetList {
		// set to low health if target was healthy before
		// reduce health even further if not detected
		// some devices seems not to respond to mdns discovery after some time,
		// so try to keep them alive here
		if d.targetList[target].Health > TargetHealthLow {
			d.targetList[target].Health = TargetHealthLow
		} else {
			// reduce health even more for each failed service discovery
			d.targetList[target].Health = d.targetList[target].Health - 1
		}
	}

	// set all discovered targets to good health
	for _, row := range targetList {
		target := row
		d.targetList[target.Address] = &target
		d.targetList[target.Address].Health = TargetHealthGood
	}

	d.logger.Debug(`finished mDNS servicediscovery"`, slog.Int("targets", len(d.targetList)))

	d.cleanup()
}

func (d *serviceDiscovery) discover(service string, timeout time.Duration, callback func(logger *slogger.Logger, target *serviceDiscoveryTarget) *DiscoveryTarget, channel chan *DiscoveryTarget) {
	wg := sync.WaitGroup{}
	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, 30)

	discoveryLogger := d.logger.With(slog.String("service", service))

	wg.Add(1)
	go func() {
		defer wg.Done()
		for entry := range entriesCh {
			target := serviceDiscoveryTarget{
				ServiceEntry: *entry,
				Name:         strings.ToLower(entry.Name),
				Host:         strings.ToLower(entry.Host),
				Port:         entry.Port,
				Address:      entry.AddrV4.String(),
				InfoFields:   map[string]string{},
				Generation:   "",
				Version:      "",
			}

			// skip if we dont have a name or address
			if target.Name == "" || target.Address == "" {
				continue
			}

			// parse info fields
			for _, field := range entry.InfoFields {
				fieldParts := strings.SplitN(field, "=", 2)
				if len(fieldParts) == 2 {
					target.InfoFields[fieldParts[0]] = fieldParts[1]
				}
			}

			if val, ok := target.InfoFields["gen"]; ok {
				target.Generation = val
			}

			if val, ok := target.InfoFields["ver"]; ok {
				target.Version = val
			}

			entryLogger := discoveryLogger.With(
				slog.Group(
					"target",
					slog.String("name", target.Name),
					slog.String("address", target.Address),
				),
			)
			if target := callback(entryLogger, &target); target != nil {
				channel <- target
			}
		}
	}()

	// Start the lookup
	params := mdns.DefaultParams(service)
	params.DisableIPv6 = true
	params.Timeout = timeout
	params.Entries = entriesCh
	params.Logger = slog.NewLogLogger(discoveryLogger.Handler(), slog.LevelInfo)
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

func (d *serviceDiscovery) SetTargetDeviceName(address, deviceName string) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if _, exists := d.targetList[address]; exists {
		d.targetList[address].DeviceName = &deviceName
	}
}

func (d *serviceDiscovery) cleanup() {
	for address, target := range d.targetList {
		if target.Health <= TargetHealthDead {
			d.logger.Debug(`disabling unhealthy target"`, slog.String("target", target.Name()), slog.String("address", target.Address))
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

func discoveryTargetFromStatic(entry string, deviceType string) DiscoveryTarget {
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
		Type:     deviceType,
		Static:   true,
		Health:   TargetHealthGood,
	}
}
