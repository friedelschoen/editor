package shadow

import (
	"image"
	"image/draw"
	"math"
)

func shadowTable(size int) []float64 {
	res := make([]float64, size)
	total := float64(size)
	for i := 0; i < size; i++ {
		// -(1/log(2))*log(x+1)+1
		res[i] = -(1/math.Log(2))*math.Log(float64(i)/total+1) + 1
	}
	return res
}

// precalculate shadow.s
var shadow = shadowTable(1000)

// maxColorDiff in [0.0, 1.0]
func PaintShadow(img draw.Image, r image.Rectangle, height int, maxColorDiff float64) {
	step := 0
	dy := float64(height)
	shades := float64(len(shadow))
	for y := r.Min.Y; y < r.Max.Y; y++ {
		yperc := float64(step) * shades / dy
		step++

		u := shadow[int(yperc)]
		v := u * maxColorDiff

		for x := r.Min.X; x < r.Max.X; x++ {
			atc := img.At(x, y)
			img.Set(x, y, Shade(atc, v))
		}
	}
}
