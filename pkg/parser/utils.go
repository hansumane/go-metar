package parser

import (
	"math"
	"strconv"
	"strings"
)

func s2i(s string) int {
	if len(s) == 0 {
		return 0
	}
	res, _ := strconv.Atoi(s)
	return res
}

func s2ui(s string) uint {
	return uint(s2i(s))
}

func unique(in []string) []string {
	var out []string
	var seen = make(map[string]bool)

	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}

	return out
}

func StringsUpper(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = strings.ToUpper(s)
	}
	return out
}

func calcHumidity(temperature, dewPoint int) float64 {
	exp := math.Exp
	tc := float64(temperature)
	tdc := float64(dewPoint)
	return 100.0 * (exp((17.625*tdc)/(243.04+tdc)) / exp((17.625*tc)/(243.04+tc)))
}
