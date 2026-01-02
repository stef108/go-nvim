package entity

import "github.com/gdamore/tcell/v2"

type Crawler struct {
	BaseEnemy
	OnGround bool
}

func NewCrawler(x, y int) *Crawler {
	return &Crawler{
		BaseEnemy: BaseEnemy{
			X:     float64(x),
			Y:     float64(y),
			VelX:  8.0,
			Char:  '🕷',
			Style: tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true),
			HP:    1, MaxHP: 1,
		},
	}
}

func (c *Crawler) Update(ctx Context) {
	c.VelY += 40.0 * ctx.DT

	// Predict next vertical position
	nextY := c.Y + c.VelY*ctx.DT

	// Ground Check (Look at feet)
	feetX := int(c.X)
	feetY := int(nextY + 0.9)

	if c.VelY > 0 && ctx.Grid.IsSolid(feetX, feetY) {
		// LANDED
		c.VelY = 0
		c.OnGround = true
		c.Y = float64(feetY - 1) // Snap to top of block
	} else {
		// FALLING
		c.OnGround = false
		c.Y = nextY
	}

	// --- MOVEMENT  ---
	if c.OnGround {
		nextX := c.X + c.VelX*ctx.DT

		wallX := int(nextX + 0.5)
		if c.VelX < 0 {
			wallX = int(nextX - 0.2)
		}

		// If hit wall -> Turn around
		if ctx.Grid.IsSolid(wallX, int(c.Y)) {
			c.VelX *= -1
		} else {
			c.X = nextX
		}

	} else {
		c.X += c.VelX * ctx.DT
	}
}
