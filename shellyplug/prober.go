package shellyplug

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

type (
	ShellyPlug struct {
		target string
		client *resty.Client
	}
)

func New(target string, client *resty.Client) *ShellyPlug {
	sp := ShellyPlug{target: target, client: client}

	return &sp
}

func (sp *ShellyPlug) GetSettings() (ResultSettings, error) {
	url := fmt.Sprintf("http://%v/settings", sp.target)

	result := ResultSettings{}

	r := sp.client.R().SetResult(&result).ForceContentType("application/json")
	_, err := r.Get(url)
	return result, err
}

func (sp *ShellyPlug) GetStatus() (ResultStatus, error) {
	url := fmt.Sprintf("http://%v/status", sp.target)

	result := ResultStatus{}

	r := sp.client.R().SetResult(&result).ForceContentType("application/json")
	_, err := r.Get(url)
	return result, err
}

func (sp *ShellyPlug) GetPowerUsage() (ResultPowerUsage, error) {
	url := fmt.Sprintf("http://%v/meter/0", sp.target)

	result := ResultPowerUsage{}

	r := sp.client.R().SetResult(&result).ForceContentType("application/json")
	_, err := r.Get(url)
	return result, err
}
