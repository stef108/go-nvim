package entity

import (
	"neovim-game/internal/core/grid"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// BaseEnemy holds the common properties for all baddies
type BaseEnemy struct {
	X, Y       float64
	VelX, VelY float64
	Char       rune
	Style      tcell.Style
	HP         int
	MaxHP      int
}

func (b *BaseEnemy) RenderPos() (int, int) {
	return int(b.X + 0.5), int(b.Y)
}

func (e *BaseEnemy) IsHit(x, y int) bool {
	ey := int(e.Y + 0.5)
	if ey != y {
		return false
	}

	ex := int(e.X + 0.5)

	// Calculate how wide this enemy actually is
	width := runewidth.RuneWidth(e.Char)

	return x >= ex && x < ex+width
}

// Render draws the entity directly to the grid
func (b *BaseEnemy) Render(g *grid.Grid) {
	x, y := int(b.X+0.5), int(b.Y)
	if x >= 0 && x < g.Width && y >= 0 && y < g.Height {
		g.SetContent(x, y, b.Char, b.Style)
	}
}

func (b *BaseEnemy) Position() (float64, float64) {
	return b.X, b.Y
}

func (b *BaseEnemy) TakeDamage(amount int) bool {
	b.HP -= amount
	return b.HP <= 0
}

func (b *BaseEnemy) IsArmored() bool {
	return false
}
