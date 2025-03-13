package discovery

import (
	"fmt"
	"strings"
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
	}
)

func (t *DiscoveryTarget) Name() string {
	return fmt.Sprintf(`%v [%v]`, t.Hostname, t.Address)
}

func (t *DiscoveryTarget) Url(path string) string {
	return fmt.Sprintf("http://%v:%v/%s", t.Address, t.Port, strings.TrimLeft(path, "/"))
}
