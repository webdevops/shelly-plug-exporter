package shellyprober

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/patrickmn/go-cache"

	"github.com/webdevops/shelly-plug-exporter/discovery"
)

type (
	ShellyProberGen2 struct {
		Target discovery.DiscoveryTarget
		Client *resty.Client
		Ctx    context.Context
		Cache  *cache.Cache
	}

	ShellyProberGen2ResultShellyConfig map[string]interface{}

	ShellyProberGen2ResultSysStatus struct {
		Mac              string `json:"mac"`
		RestartRequired  bool   `json:"restart_required"`
		Time             string `json:"time"`
		Unixtime         int    `json:"unixtime"`
		Uptime           int    `json:"uptime"`
		RAMSize          int    `json:"ram_size"`
		RAMFree          int    `json:"ram_free"`
		FsSize           int    `json:"fs_size"`
		FsFree           int    `json:"fs_free"`
		CfgRev           int    `json:"cfg_rev"`
		KvsRev           int    `json:"kvs_rev"`
		ScheduleRev      int    `json:"schedule_rev"`
		WebhookRev       int    `json:"webhook_rev"`
		AvailableUpdates struct {
			Stable struct {
				Version string `json:"version"`
			} `json:"stable"`
		} `json:"available_updates"`
	}

	ShellyProberGen2ResultWifiStatus struct {
		StaIP  string `json:"sta_ip"`
		Status string `json:"status"`
		Ssid   string `json:"ssid"`
		Rssi   int    `json:"rssi"`
	}

	ShellyProberGen2ResultTemperature struct {
		ID int     `json:"id"`
		TC float64 `json:"tC"`
		TF float64 `json:"tF"`
	}

	ShellyProberGen2ResultSwitch struct {
		ID      int     `json:"id"`
		Source  string  `json:"source"`
		Output  bool    `json:"output"`
		Apower  float64 `json:"apower"`
		Voltage float64 `json:"voltage"`
		Current float64 `json:"current"`
		Aenergy struct {
			Total    float64   `json:"total"`
			ByMinute []float64 `json:"by_minute"`
			MinuteTs float64   `json:"minute_ts"`
		} `json:"aenergy"`
		Temperature struct {
			TC float64 `json:"tC"`
			TF float64 `json:"tF"`
		} `json:"temperature"`
	}

	ShellyProberGen2ResultEm struct {
		ID                  int     `json:"id"`
		ACurrent            float64 `json:"a_current"`
		AVoltage            float64 `json:"a_voltage"`
		AActPower           float64 `json:"a_act_power"`
		AAprtPower          float64 `json:"a_aprt_power"`
		APf                 float64 `json:"a_pf"`
		AFreq               float64 `json:"a_freq"`
		BCurrent            float64 `json:"b_current"`
		BVoltage            float64 `json:"b_voltage"`
		BActPower           float64 `json:"b_act_power"`
		BAprtPower          float64 `json:"b_aprt_power"`
		BPf                 float64 `json:"b_pf"`
		BFreq               float64 `json:"b_freq"`
		CCurrent            float64 `json:"c_current"`
		CVoltage            float64 `json:"c_voltage"`
		CActPower           float64 `json:"c_act_power"`
		CAprtPower          float64 `json:"c_aprt_power"`
		CPf                 float64 `json:"c_pf"`
		CFreq               float64 `json:"c_freq"`
		NCurrent            any     `json:"n_current"`
		TotalCurrent        float64 `json:"total_current"`
		TotalActPower       float64 `json:"total_act_power"`
		TotalAprtPower      float64 `json:"total_aprt_power"`
		UserCalibratedPhase []any   `json:"user_calibrated_phase"`
	}

	ShellyProberGen2ResultEmData struct {
		ID                 int     `json:"id"`
		ATotalActEnergy    float64 `json:"a_total_act_energy"`
		ATotalActRetEnergy float64 `json:"a_total_act_ret_energy"`
		BTotalActEnergy    float64 `json:"b_total_act_energy"`
		BTotalActRetEnergy float64 `json:"b_total_act_ret_energy"`
		CTotalActEnergy    float64 `json:"c_total_act_energy"`
		CTotalActRetEnergy float64 `json:"c_total_act_ret_energy"`
		TotalAct           float64 `json:"total_act"`
		TotalActRet        float64 `json:"total_act_ret"`
	}
)

func (sp *ShellyProberGen2) fetch(url string, target interface{}) error {
	r := sp.Client.R().SetContext(sp.Ctx).SetResult(&target).ForceContentType("application/json")
	_, err := r.Get(sp.Target.Url(url))
	return err
}

func (sp *ShellyProberGen2) fetchWithCache(url string, target interface{}) error {
	url = sp.Target.Url(url)
	cacheKey := url

	if val, ok := sp.Cache.Get(cacheKey); ok {
		if data, err := json.Marshal(val); err == nil {
			if err := json.Unmarshal(data, target); err == nil {
				return nil
			}
		}
	}

	r := sp.Client.R().SetContext(sp.Ctx).SetResult(&target).ForceContentType("application/json")
	_, err := r.Get(url)

	sp.Cache.SetDefault(cacheKey, target)

	return err
}

func (sp *ShellyProberGen2) GetSysStatus() (ShellyProberGen2ResultSysStatus, error) {
	result := ShellyProberGen2ResultSysStatus{}
	err := sp.fetch("/rpc/Sys.GetStatus", &result)
	return result, err
}

func (sp *ShellyProberGen2) GetShellyConfig() (ShellyProberGen2ResultShellyConfig, error) {
	result := ShellyProberGen2ResultShellyConfig{}
	err := sp.fetchWithCache("/rpc/Shelly.GetConfig", &result)
	return result, err
}

func (sp *ShellyProberGen2) GetWifiStatus() (ShellyProberGen2ResultWifiStatus, error) {
	result := ShellyProberGen2ResultWifiStatus{}
	err := sp.fetch("/rpc/Wifi.GetStatus", &result)
	return result, err
}

func (sp *ShellyProberGen2) GetTemperatureStatus(id int) (ShellyProberGen2ResultTemperature, error) {
	result := ShellyProberGen2ResultTemperature{}
	err := sp.fetch(fmt.Sprintf("/rpc/Temperature.GetStatus?id=%d", id), &result)
	return result, err
}

func (sp *ShellyProberGen2) GetSwitchStatus(id int) (ShellyProberGen2ResultSwitch, error) {
	result := ShellyProberGen2ResultSwitch{}
	err := sp.fetch(fmt.Sprintf("/rpc/Switch.GetStatus?id=%d", id), &result)
	return result, err
}

func (sp *ShellyProberGen2) GetEmStatus(id int) (ShellyProberGen2ResultEm, error) {
	result := ShellyProberGen2ResultEm{}
	err := sp.fetch(fmt.Sprintf("/rpc/Em.GetStatus?id=%d", id), &result)
	return result, err
}

func (sp *ShellyProberGen2) GetEmDataStatus(id int) (ShellyProberGen2ResultEmData, error) {
	result := ShellyProberGen2ResultEmData{}
	err := sp.fetch(fmt.Sprintf("/rpc/EmData.GetStatus?id=%d", id), &result)
	return result, err
}
