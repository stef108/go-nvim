package engine

import (
	"sync"

	"neovim-game/internal/core/entity"
	"neovim-game/internal/core/grid"
)

type Engine struct {
	// State
	Enemies   []*entity.Enemy
	Particles []*entity.Particle

	// The Buffers
	CodeGrid   *grid.Grid // The Clean state from Neovim
	RenderGrid *grid.Grid // The altered state sent to TUI

	stopChan chan struct{}
	mu       sync.RWMutex
}

func New(w, h int) *Engine {
	return &Engine{
		Enemies:    make([]*entity.Enemy, 0),
		Particles:  make([]*entity.Particle, 0),
		CodeGrid:   grid.New(w, h),
		RenderGrid: grid.New(w, h),
		stopChan:   make(chan struct{}),
	}
}

func (e *Engine) Resize(w, h int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.CodeGrid.Resize(w, h)
	e.RenderGrid.Resize(w, h)
}

func (e *Engine) SpawnEnemy(x, y int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.Enemies = append(e.Enemies, entity.NewBug(x, y))
}

func (e *Engine) Explode(x, y float64) {
	debris := entity.NewExplosion(x, y)
	e.Particles = append(e.Particles, debris...)
}

// Update runs physics (independent of render)
func (e *Engine) Update(dt float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	cursorX := e.CodeGrid.CursorX
	cursorY := e.CodeGrid.CursorY
	// Update all enemies
	activeEnemies := e.Enemies[:0]

	for _, bug := range e.Enemies {
		bug.Update(dt)

		if bug.IsHit(cursorX, cursorY) {
			e.Explode(bug.X, bug.Y)
			continue
		}
		// Cast Height to float64 for comparison
		if bug.Y < float64(e.CodeGrid.Height) {
			activeEnemies = append(activeEnemies, bug)
		}
	}
	e.Enemies = activeEnemies

	// Update Particles
	activeParticles := e.Particles[:0]
	for _, p := range e.Particles {
		p.Update(dt)

		// Keep if alive and on screen
		if p.Lifetime > 0 && p.Y < float64(e.CodeGrid.Height) && p.Y >= 0 {
			activeParticles = append(activeParticles, p)
		}
	}
	e.Particles = activeParticles
}

// DrawCompositor combines Layers: Code + Enemies -> RenderGrid
func (e *Engine) DrawCompositor() *grid.Grid {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Copy Background (Code) to Render Buffer
	e.RenderGrid.CopyFrom(e.CodeGrid)

	// Draw Enemies on top
	for _, bug := range e.Enemies {
		// Get integer grid coordinates
		x, y := bug.RenderPos()
		// Boundary check
		if x >= 0 && x < e.RenderGrid.Width && y >= 0 && y < e.RenderGrid.Height {
			e.RenderGrid.SetContent(x, y, bug.Char, bug.Style)
		}
	}

	// Draw Particles
	for _, p := range e.Particles {
		x, y := int(p.X), int(p.Y)
		if x >= 0 && x < e.RenderGrid.Width && y >= 0 && y < e.RenderGrid.Height {
			e.RenderGrid.SetContent(x, y, p.Char, p.Style)
		}
	}

	return e.RenderGrid
}
