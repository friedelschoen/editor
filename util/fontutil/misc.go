package fontutil

import (
	"golang.org/x/image/math/fixed"
)

func Fixed266ToFloat64(v fixed.Int26_6) float64 {
	return float64(v) / float64(64)
}
