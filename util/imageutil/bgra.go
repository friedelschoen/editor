package imageutil

import (
	"image"
	"image/color"
)

type BGRA struct {
	image.RGBA
}

// Allows fast lane if detected.

func BgraColor(c color.Color) color.RGBA {
	c2 := RgbaColor(c)
	c2.R, c2.B = c2.B, c2.R // convert to BGR
	return c2
}
