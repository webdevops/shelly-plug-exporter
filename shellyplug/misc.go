package shellyplug

import (
	"github.com/prometheus/client_golang/prometheus"
)

func boolToFloat64(v bool) float64 {
	if v {
		return 1
	}

	return 0
}

func copyLabelMap(val prometheus.Labels) (ret prometheus.Labels) {
	ret = prometheus.Labels{}
	for name, val := range val {
		ret[name] = val
	}
	return
}
