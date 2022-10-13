package shellyplug

func boolToFloat64(v bool) float64 {
	if v {
		return 1
	}

	return 0
}
