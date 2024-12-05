package exchange

import "strconv"

// parseFloat64 converts a string to float64, returns 0 if error
func parseFloat64(val string) float64 {
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0
	}
	return f
}
