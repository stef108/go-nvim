package entity

import (
	"neovim-game/internal/core/grid"
)

// Context contains external world data an entity needs to update
type Context struct {
	DT      float64
	Grid    *grid.Grid
	PlayerX int
	PlayerY int
}

// GameEntity is the interface for ANYTHING that lives in the game world
type GameEntity interface {
	// Lifecycle
	Update(ctx Context)
	Render(g *grid.Grid)

	// Combat
	IsHit(x, y int) bool
	TakeDamage(amount int) (dead bool)

	// Data Access (For the Engine to sort/cull)
	Position() (float64, float64)
	IsArmored() bool
}
