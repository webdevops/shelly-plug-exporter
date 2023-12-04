package shellyprober

import (
	"context"

	"github.com/go-resty/resty/v2"
	"github.com/patrickmn/go-cache"
	"github.com/webdevops/shelly-plug-exporter/discovery"
)

type (
	ShellyProberGen1Em struct {
		Target discovery.DiscoveryTarget
		Client *resty.Client
		Ctx    context.Context
		Cache  *cache.Cache
	}

	ShellyProberGen1EmResultSettings struct {
		Name     string  `json:"name"`
		MaxPower float64 `json:"max_power"`
		Fw       string  `json:"fw"`

		Device struct {
			Hostname string `json:"hostname"`
			Mac      string `json:"mac"`
			Type     string `json:"type"`
		} `json:"device"`
	}

	ShellyProberGen1EmResultStatus struct {
		WifiSta struct {
			Connected bool   `yaml:"connected"`
			Ssid      string `yaml:"ssid"`
			IP        string `yaml:"ip"`
			Rssi      int    `yaml:"rssi"`
		} `yaml:"wifi_sta"`
		Cloud struct {
			Enabled   bool `yaml:"enabled"`
			Connected bool `yaml:"connected"`
		} `yaml:"cloud"`
		Mqtt struct {
			Connected bool `yaml:"connected"`
		} `yaml:"mqtt"`
		Time          string `yaml:"time"`
		Unixtime      int    `yaml:"unixtime"`
		Serial        int    `yaml:"serial"`
		HasUpdate     bool   `yaml:"has_update"`
		Mac           string `yaml:"mac"`
		CfgChangedCnt int    `yaml:"cfg_changed_cnt"`
		ActionsStats  struct {
			Skipped int `yaml:"skipped"`
		} `yaml:"actions_stats"`
		Relays []struct {
			Ison           bool   `yaml:"ison"`
			HasTimer       bool   `yaml:"has_timer"`
			TimerStarted   int    `yaml:"timer_started"`
			TimerDuration  int    `yaml:"timer_duration"`
			TimerRemaining int    `yaml:"timer_remaining"`
			Overpower      bool   `yaml:"overpower"`
			IsValid        bool   `yaml:"is_valid"`
			Source         string `yaml:"source"`
		} `yaml:"relays"`
		Emeters []struct {
			Power         float64 `yaml:"power"`
			Pf            float64 `yaml:"pf"`
			Current       float64 `yaml:"current"`
			Voltage       float64 `yaml:"voltage"`
			IsValid       bool    `yaml:"is_valid"`
			Total         float64 `yaml:"total"`
			TotalReturned float64 `yaml:"total_returned"`
		} `yaml:"emeters"`
		TotalPower float64 `yaml:"total_power"`
		EmeterN    struct {
			Current  int     `yaml:"current"`
			Ixsum    float64 `yaml:"ixsum"`
			Mismatch bool    `yaml:"mismatch"`
			IsValid  bool    `yaml:"is_valid"`
		} `yaml:"emeter_n"`
		FsMounted bool `yaml:"fs_mounted"`
		VData     int  `yaml:"v_data"`
		CtCalst   int  `yaml:"ct_calst"`
		Update    struct {
			Status      string `yaml:"status"`
			HasUpdate   bool   `yaml:"has_update"`
			NewVersion  string `yaml:"new_version"`
			OldVersion  string `yaml:"old_version"`
			BetaVersion string `yaml:"beta_version"`
		} `yaml:"update"`
		RAMTotal int `yaml:"ram_total"`
		RAMFree  int `yaml:"ram_free"`
		FsSize   int `yaml:"fs_size"`
		FsFree   int `yaml:"fs_free"`
		Uptime   int `yaml:"uptime"`
	}
)

func (sp *ShellyProberGen1Em) fetch(url string, target interface{}) error {
	r := sp.Client.R().SetContext(sp.Ctx).SetResult(&target).ForceContentType("application/json")
	_, err := r.Get(sp.Target.Url(url))
	return err
}

func (sp *ShellyProberGen1Em) GetSettings() (ShellyProberGen1EmResultSettings, error) {
	result := ShellyProberGen1EmResultSettings{}
	err := sp.fetch("/settings", &result)
	return result, err
}

func (sp *ShellyProberGen1Em) GetStatus() (ShellyProberGen1EmResultStatus, error) {
	result := ShellyProberGen1EmResultStatus{}
	err := sp.fetch("/status", &result)
	return result, err
}
