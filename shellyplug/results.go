package shellyplug

import (
	"github.com/webdevops/shelly-plug-exporter/discovery"
)

type (
	ResultShellyInfo struct {
		Name       string      `json:"name"`
		ID         string      `json:"id"`
		Mac        string      `json:"mac"`
		Model      string      `json:"model"`
		Gen        *int        `json:"gen"`
		FwID       string      `json:"fw_id"`
		Ver        string      `json:"ver"`
		App        string      `json:"app"`
		AuthEn     bool        `json:"auth_en"`
		AuthDomain interface{} `json:"auth_domain"`
	}
)

func (sp *ShellyPlug) targetGetShellyInfo(target discovery.DiscoveryTarget) (ResultShellyInfo, error) {
	result := ResultShellyInfo{}

	r := sp.client.R().SetContext(sp.ctx).SetResult(&result).ForceContentType("application/json")
	_, err := r.Get(target.Url("/shelly"))
	return result, err
}
