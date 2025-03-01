package drawer

import "image"

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
