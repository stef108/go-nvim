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
func NewExplosion(x, y float64, intensity int) []*Particle {
	if intensity < 1 {
		intensity = 1
	}
	if intensity > 50 {
		intensity = 50
	}

	// Base of 5 particles, plus more for higher intensity
	count := 5 + (intensity / 2)

	// Higher intensity = debris flies faster and further
	speedMult := 1.0 + (float64(intensity) * 0.1)

	particles := make([]*Particle, count)

	for i := range count {
		// Random angle
		vx := (rand.Float64() - 0.5) * 20.0 * speedMult
		vy := (rand.Float64() - 0.5) * 10.0 * speedMult

		chars := []rune{'*', '.', ',', 'o', 'x'}
		c := chars[rand.Intn(len(chars))]

		style := tcell.StyleDefault.Foreground(tcell.ColorOrange)
		if intensity > 15 {
			style = tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
		}

		particles[i] = &Particle{
			X:        x,
			Y:        y,
			VelX:     vx,
			VelY:     vy,
			Char:     c,
			Style:    style,
			Lifetime: 0.5 + (rand.Float64() * 0.5), // Random lifetime 0.5-1.0s
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
