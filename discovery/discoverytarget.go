package discovery

import (
	"fmt"
	"strings"
)

type (
	DiscoveryTarget struct {
		Hostname string
		Address  string
		Port     int
		Health   int
		Type     string
		Static   bool
	}
)

func (t *DiscoveryTarget) Name() string {
	return fmt.Sprintf(`%v [%v]`, t.Hostname, t.Address)
}

func (t *DiscoveryTarget) Url(path string) string {
	return fmt.Sprintf("http://%v:%v/%s", t.Address, t.Port, strings.TrimLeft(path, "/"))
}
