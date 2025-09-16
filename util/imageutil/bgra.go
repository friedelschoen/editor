package imageutil

import (
	"image"
	"image/color"
)

// BGRA is an in-memory image whose At method returns [color.BGRA] values.
type BGRA struct {
	// Pix holds the image's pixels, in R, G, B, A order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*4].
	Pix []uint8
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
}

func (p *BGRA) ColorModel() color.Model { return color.RGBAModel }

func (p *BGRA) Bounds() image.Rectangle { return p.Rect }

func (p *BGRA) At(x, y int) color.Color {
	return p.RGBAAt(x, y)
}

func (p *BGRA) RGBAAt(x, y int) color.RGBA {
	if !(image.Point{x, y}.In(p.Rect)) {
		return color.RGBA{}
	}
	i := p.PixOffset(x, y)
	s := p.Pix[i : i+4 : i+4] // Small cap improves performance, see https://golang.org/issue/27857
	return color.RGBA{s[0], s[1], s[2], s[3]}
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (p *BGRA) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*4
}

func (p *BGRA) RGBA64At(x, y int) color.RGBA64 {
	if !(image.Point{x, y}.In(p.Rect)) {
		return color.RGBA64{}
	}
	i := p.PixOffset(x, y)
	s := p.Pix[i : i+4 : i+4] // Small cap improves performance, see https://golang.org/issue/27857
	r := uint16(s[0])
	g := uint16(s[1])
	b := uint16(s[2])
	a := uint16(s[3])
	return color.RGBA64{
		(r << 8) | r,
		(g << 8) | g,
		(b << 8) | b,
		(a << 8) | a,
	}
}

func (p *BGRA) Set(x, y int, c color.Color) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)
	c1 := color.RGBAModel.Convert(c).(color.RGBA)
	s := p.Pix[i : i+4 : i+4] // Small cap improves performance, see https://golang.org/issue/27857
	s[2] = c1.R
	s[1] = c1.G
	s[0] = c1.B
	s[3] = c1.A
}

func (p *BGRA) SetRGBA64(x, y int, c color.RGBA64) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)
	s := p.Pix[i : i+4 : i+4] // Small cap improves performance, see https://golang.org/issue/27857
	s[2] = uint8(c.R >> 8)
	s[1] = uint8(c.G >> 8)
	s[0] = uint8(c.B >> 8)
	s[3] = uint8(c.A >> 8)
}

func (p *BGRA) SetRGBA(x, y int, c color.RGBA) {
	if !(image.Point{x, y}.In(p.Rect)) {
		return
	}
	i := p.PixOffset(x, y)
	s := p.Pix[i : i+4 : i+4] // Small cap improves performance, see https://golang.org/issue/27857
	s[0] = c.R
	s[1] = c.G
	s[2] = c.B
	s[3] = c.A
}

// SubImage returns an image representing the portion of the image p visible
// through r. The returned value shares pixels with the original image.
func (p *BGRA) SubImage(r image.Rectangle) image.Image {
	r = r.Intersect(p.Rect)
	// If r1 and r2 are Rectangles, r1.Intersect(r2) is not guaranteed to be inside
	// either r1 or r2 if the intersection is empty. Without explicitly checking for
	// this, the Pix[i:] expression below can panic.
	if r.Empty() {
		return &BGRA{}
	}
	i := p.PixOffset(r.Min.X, r.Min.Y)
	return &BGRA{
		Pix:    p.Pix[i:],
		Stride: p.Stride,
		Rect:   r,
	}
}

// Opaque scans the entire image and reports whether it is fully opaque.
func (p *BGRA) Opaque() bool {
	if p.Rect.Empty() {
		return true
	}
	i0, i1 := 3, p.Rect.Dx()*4
	for y := p.Rect.Min.Y; y < p.Rect.Max.Y; y++ {
		for i := i0; i < i1; i += 4 {
			if p.Pix[i] != 0xff {
				return false
			}
		}
		i0 += p.Stride
		i1 += p.Stride
	}
	return true
}

// NewBGRA returns a new [BGRA] image with the given bounds.
func NewBGRA(r image.Rectangle) *BGRA {
	return &BGRA{
		Pix:    make([]uint8, 4*r.Dx()*r.Dy()),
		Stride: 4 * r.Dx(),
		Rect:   r,
	}
}

// NewBGRA returns a new [BGRA] image with the given bounds.
func NewBGRAFromBuffer(buf []byte, r image.Rectangle) *BGRA {
	if len(buf) != 4*r.Dx()*r.Dy() {
		return nil
	}
	return &BGRA{
		Pix:    buf,
		Stride: 4 * r.Dx(),
		Rect:   r,
	}
}
