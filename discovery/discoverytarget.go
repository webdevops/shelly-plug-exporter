package discovery

import (
	"fmt"
)

type (
	DiscoveryTarget struct {
		DeviceName *string `json:"deviceName"`
		Hostname   string  `json:"hostname"`
		Address    string  `json:"address"`
		Port       int     `json:"port"`
		Health     int     `json:"health"`
		Type       string  `json:"type"`
		Static     bool    `json:"isStatic"`
		Generation string  `json:"generation"`
	}
)

func (t *DiscoveryTarget) Name() string {
	return fmt.Sprintf(`%v [%v]`, t.Hostname, t.Address)
}

func (t *DiscoveryTarget) BaseUrl() string {
	if t.Port == 80 {
		return fmt.Sprintf("http://%v", t.Address)
	} else {
		return fmt.Sprintf("http://%v:%v", t.Address, t.Port)
	}
}
