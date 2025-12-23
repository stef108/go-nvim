package entity

import (
	"math/rand"

	"github.com/gdamore/tcell/v2"
)

type Particle struct {
	X, Y       float64
	VelX, VelY float64
	Char       rune
	Style      tcell.Style
	Lifetime   float64 // Secs until it disappears
}

// NewExplosion creates a cluster of particles at a point
func NewExplosion(x, y float64) []*Particle {
	count := 8 // Number of debris pieces
	particles := make([]*Particle, count)

	for i := 0; i < count; i++ {
		// Random angle and speed
		// VelX: -10 to 10
		// VelY: -5 to 5
		vx := (rand.Float64() - 0.5) * 20.0
		vy := (rand.Float64() - 0.5) * 10.0

		// Random character for debris
		chars := []rune{'*', '.', ',', 'o', 'x'}
		c := chars[rand.Intn(len(chars))]

		particles[i] = &Particle{
			X:        x,
			Y:        y,
			VelX:     vx,
			VelY:     vy,
			Char:     c,
			Style:    tcell.StyleDefault.Foreground(tcell.ColorOrange).Bold(false),
			Lifetime: 0.8,
		}
	}
	return particles
}

func (p *Particle) Update(dt float64) {
	p.Lifetime -= dt

	// Physics
	p.X += p.VelX * dt
	p.Y += p.VelY * dt

	// Gravity (pull down)
	p.VelY += 15.0 * dt

	// Drag (air resistance)
	p.VelX *= 0.95
	p.VelY *= 0.95
}
