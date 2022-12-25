package shellyprober

import (
	"context"

	"github.com/go-resty/resty/v2"
	"github.com/patrickmn/go-cache"

	"github.com/webdevops/shelly-plug-exporter/discovery"
)

type (
	ShellyProberGen1 struct {
		Target discovery.DiscoveryTarget
		Client *resty.Client
		Ctx    context.Context
		Cache  *cache.Cache
	}

	ShellyProberGen1ResultSettings struct {
		Name     string  `json:"name"`
		MaxPower float64 `json:"max_power"`
		Fw       string  `json:"fw"`

		Device struct {
			Hostname string `json:"hostname"`
			Mac      string `json:"mac"`
			Type     string `json:"type"`
		} `json:"device"`
	}

	ShellyProberGen1ResultStatus struct {
		WifiSta struct {
			Connected bool   `json:"connected"`
			Ssid      string `json:"ssid"`
			IP        string `json:"ip"`
			Rssi      int    `json:"rssi"`
		} `json:"wifi_sta"`
		Cloud struct {
			Enabled   bool `json:"enabled"`
			Connected bool `json:"connected"`
		} `json:"cloud"`
		Mqtt struct {
			Connected bool `json:"connected"`
		} `json:"mqtt"`
		Time          string `json:"time"`
		Unixtime      int    `json:"unixtime"`
		Serial        int    `json:"serial"`
		HasUpdate     bool   `json:"has_update"`
		Mac           string `json:"mac"`
		CfgChangedCnt int    `json:"cfg_changed_cnt"`
		ActionsStats  struct {
			Skipped int `json:"skipped"`
		} `json:"actions_stats"`
		Relays []struct {
			Ison           bool   `json:"ison"`
			HasTimer       bool   `json:"has_timer"`
			TimerStarted   int    `json:"timer_started"`
			TimerDuration  int    `json:"timer_duration"`
			TimerRemaining int    `json:"timer_remaining"`
			Overpower      bool   `json:"overpower"`
			Source         string `json:"source"`
		} `json:"relays"`
		Meters []struct {
			Power     float64   `json:"power"`
			Overpower float64   `json:"overpower"`
			IsValid   bool      `json:"is_valid"`
			Timestamp int       `json:"timestamp"`
			Counters  []float64 `json:"counters"`
			Total     float64   `json:"total"`
		} `json:"meters"`
		Temperature     float64 `json:"temperature"`
		Overtemperature bool    `json:"overtemperature"`
		Tmp             struct {
			TC      float64 `json:"tC"`
			TF      float64 `json:"tF"`
			IsValid bool    `json:"is_valid"`
		} `json:"tmp"`
		Update struct {
			Status     string `json:"status"`
			HasUpdate  bool   `json:"has_update"`
			NewVersion string `json:"new_version"`
			OldVersion string `json:"old_version"`
		} `json:"update"`
		RAMTotal int `json:"ram_total"`
		RAMFree  int `json:"ram_free"`
		FsSize   int `json:"fs_size"`
		FsFree   int `json:"fs_free"`
		Uptime   int `json:"uptime"`
	}

	ShellyProberGen1ResultPowerUsage struct {
		Power     float64   `json:"power"`
		Overpower float64   `json:"overpower"`
		IsValid   bool      `json:"is_valid"`
		Timestamp int       `json:"timestamp"`
		Counters  []float64 `json:"counters"`
		Total     int       `json:"total"`
	}
)

func (sp *ShellyProberGen1) fetch(url string, target interface{}) error {
	r := sp.Client.R().SetContext(sp.Ctx).SetResult(&target).ForceContentType("application/json")
	_, err := r.Get(sp.Target.Url(url))
	return err
}

func (sp *ShellyProberGen1) GetSettings() (ShellyProberGen1ResultSettings, error) {
	result := ShellyProberGen1ResultSettings{}
	err := sp.fetch("/settings", &result)
	return result, err
}

func (sp *ShellyProberGen1) GetStatus() (ShellyProberGen1ResultStatus, error) {
	result := ShellyProberGen1ResultStatus{}
	err := sp.fetch("/status", &result)
	return result, err
}
