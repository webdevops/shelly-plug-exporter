package shellyplug

import (
	"github.com/webdevops/shelly-plug-exporter/discovery"
)

func (sp *ShellyPlug) UseDiscovery() {
	if discovery.ServiceDiscovery == nil {
		panic(`servicediscovery not enabled`)
	}

	sp.SetTargets(discovery.ServiceDiscovery.GetTargetList())
}
