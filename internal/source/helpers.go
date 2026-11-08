package source

import "strconv"

func parseFloat(value string) float64 {
	number, _ := strconv.ParseFloat(value, 64)
	return number
}
