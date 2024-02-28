package ctypes

type Position struct {
	x float64
	y float64
}

func NewPosition(x, y float64) Position {
	return Position{x: x, y: y}
}

// AffectX affects x in the rightwards (positive) direction, as ebiten's origin (0,0) is
// at the top left of the window.
func (p *Position) AffectX(dx float64) {
	p.x += dx
}

// AffectY affects y in the downwards (negative) direction, as ebiten's origin (0,0) is
// at the top left of the window.
func (p *Position) AffectY(dy float64) {
	p.y -= dy
}

func (p *Position) X() float64 {
	return p.x
}

func (p *Position) Y() float64 {
	return p.y
}
