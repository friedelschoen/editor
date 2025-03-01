package imageutil

import (
	"image"
	"image/color"
	"math"

	"github.com/friedelschoen/glake/util/mathutil"
)

func MaxPoint(p1, p2 image.Point) image.Point {
	if p1.X < p2.X {
		p1.X = p2.X
	}
	if p1.Y < p2.Y {
		p1.Y = p2.Y
	}
	return p1
}
func MinPoint(p1, p2 image.Point) image.Point {
	if p1.X > p2.X {
		p1.X = p2.X
	}
	if p1.Y > p2.Y {
		p1.Y = p2.Y
	}
	return p1
}

func RgbaColor(c color.Color) color.RGBA {
	if u, ok := c.(color.RGBA); ok {
		return u
	} else {
		r, g, b, a := c.RGBA()
		return color.RGBA{
			uint8(r >> 8),
			uint8(g >> 8),
			uint8(b >> 8),
			uint8(a >> 8),
		}
	}
}

func RgbaFromInt(u int) color.RGBA {
	v := u & 0xffffff
	r := uint8((v << 0) >> 16)
	g := uint8((v << 8) >> 16)
	b := uint8((v << 16) >> 16)
	return color.RGBA{r, g, b, 255}
}
func RgbaToInt(c color.RGBA) int {
	v := 0
	v += int(c.R) << 16
	v += int(c.G) << 8
	v += int(c.B) << 0
	return v
}

// NOTE: https://www.pyimagesearch.com/2015/10/05/opencv-gamma-correction/
func NewLinearInvertFn(v1, v2 float64) func(color.RGBA) color.RGBA {
	gt1 := NewGammaTable(v1)
	gt2 := NewGammaTable(v2)
	return func(c color.RGBA) color.RGBA {
		c = gt1.Lookup(c)
		// bitwise not
		c.R = ^c.R
		c.G = ^c.G
		c.B = ^c.B

		return gt2.Lookup(c)
	}
}
func NewLinearInvertFn2(v1, v2 float64) func(color.Color) color.Color {
	fn := NewLinearInvertFn(v1, v2)
	return func(c color.Color) color.Color {
		return fn(RgbaColor(c))
	}
}

type GammaTable struct {
	Table [256]uint8
}

func NewGammaTable(gamma float64) *GammaTable {
	g := &GammaTable{}
	gamma = math.Max(0.00001, gamma)
	for i := 0; i < 256; i++ {
		g.Table[i] = uint8(mathutil.Limit(
			math.Pow(float64(i)/255, 1.0/gamma)*255,
			0, 255,
		))
	}
	return g
}
func (g *GammaTable) Lookup(c color.RGBA) color.RGBA {
	return color.RGBA{
		g.Table[c.R],
		g.Table[c.G],
		g.Table[c.B],
		c.A,
	}
}
