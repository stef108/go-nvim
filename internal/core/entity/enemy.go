package entity

import (
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

type Enemy struct {
	ID    int
	X, Y  float64
	Char  rune
	Style tcell.Style
	// Velocity (Cells per second)
	VelX, VelY float64
	IsDead     bool
}

func NewBug(x, y int) *Enemy {
	return &Enemy{
		ID:    rand.Int(),
		X:     float64(x),
		Y:     float64(y),
		Char:  '👾',
		Style: tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true),

		// Fall speed
		VelY: 8,
		VelX: 0,
	}
}

// Update uses Delta Time (dt) for sub-cell movement
func (e *Enemy) Update(dt float64) {
	e.X += e.VelX * dt
	e.Y += e.VelY * dt
}

// RenderPos returns the Grid integer coordinates
func (e *Enemy) RenderPos() (int, int) {
	// Simple truncation (4.9 -> 4)
	return int(e.X), int(e.Y)
}

func (e *Enemy) IsHit(x, y int) bool {
	return int(e.X) == x && int(e.Y) == y
}
