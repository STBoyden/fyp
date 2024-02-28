package ctypes

type Position struct {
	X float64
	Y float64
}

func NewPosition(x, y float64) Position {
	return Position{X: x, Y: y}
}

// AffectX affects x in the rightwards (positive) direction, as ebiten's origin (0,0) is
// at the top left of the window.
func (p *Position) AffectX(dx float64) {
	p.X += dx
}

// AffectY affects y in the downwards (negative) direction, as ebiten's origin (0,0) is
// at the top left of the window.
func (p *Position) AffectY(dy float64) {
	p.Y -= dy
}
