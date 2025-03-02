package mathutil

import (
	"cmp"
	"math"
)

func RoundFloat64(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

// TODO: remove
func LimitFloat64(v float64, min, max float64) float64 {
	if v < min {
		return min
	} else if v > max {
		return max
	}
	return v
}

func Min[T cmp.Ordered](s ...T) T {
	m := s[0]
	for _, v := range s[1:] {
		if m > v {
			m = v
		}
	}
	return m
}

func Max[T cmp.Ordered](s ...T) T {
	m := s[0]
	for _, v := range s[1:] {
		if m < v {
			m = v
		}
	}
	return m
}
