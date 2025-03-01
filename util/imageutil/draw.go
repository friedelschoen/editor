package imageutil

import (
	"image"
	"image/color"
	"image/draw"
)

func DrawMask(
	dst draw.Image,
	r image.Rectangle,
	src image.Image, srcp image.Point,
	mask image.Image, maskp image.Point,
	op draw.Op,
) {
	// improve performance for bgra
	if bgra, ok := dst.(*BGRA); ok {
		dst = &bgra.RGBA
	}

	draw.DrawMask(dst, r, src, srcp, mask, maskp, op)
}

func DrawUniformMask(
	dst draw.Image,
	r image.Rectangle,
	c color.Color,
	mask image.Image, maskp image.Point,
	op draw.Op,
) {
	if c == nil {
		return
	}
	// correct color for bgra
	if _, ok := dst.(*BGRA); ok {
		c = BgraColor(c)
	}

	src := image.NewUniform(c)
	srcp := image.Point{}
	DrawMask(dst, r, src, srcp, mask, maskp, op)
}

func DrawUniform(dst draw.Image, r image.Rectangle, c color.Color, op draw.Op) {
	DrawUniformMask(dst, r, c, nil, image.Point{}, op)
}

func FillRectangle(img draw.Image, r image.Rectangle, c color.Color) {
	DrawUniform(img, r, c, draw.Src)
}
