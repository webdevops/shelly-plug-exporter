package shellyplug

type (
	ResultSettings struct {
		Name     string  `json:"name"`
		MaxPower float64 `json:"max_power"`
		Fw       string  `json:"fw"`

		Device struct {
			Hostname string `json:"hostname"`
			Mac      string `json:"mac"`
			Type     string `json:"type"`
		} `json:"device"`
	}

	ResultStatus struct {
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

	ResultPowerUsage struct {
		Power     float64   `json:"power"`
		Overpower float64   `json:"overpower"`
		IsValid   bool      `json:"is_valid"`
		Timestamp int       `json:"timestamp"`
		Counters  []float64 `json:"counters"`
		Total     int       `json:"total"`
	}
)
