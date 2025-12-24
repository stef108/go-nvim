package engine

import (
	"fmt"
	"math/rand"
	"sync"

	"neovim-game/internal/core/entity"
	"neovim-game/internal/core/grid"

	"github.com/gdamore/tcell/v2"
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

	// Game Rules
	Score      int
	Health     int
	MaxHealth  int
	IsGameOver bool

	// Wave Logic
	SpawnTimer float64
	SpawnRate  float64
}

func New(w, h int) *Engine {
	return &Engine{
		Enemies:    make([]*entity.Enemy, 0),
		Particles:  make([]*entity.Particle, 0),
		CodeGrid:   grid.New(w, h),
		RenderGrid: grid.New(w, h),
		stopChan:   make(chan struct{}),

		// Defaults
		Score:      0,
		Health:     100,
		MaxHealth:  100,
		SpawnRate:  2.0,
		SpawnTimer: 0,
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

	if e.IsGameOver {
		return
	}

	// WAVE SPAWNER ---
	e.SpawnTimer -= dt
	if e.SpawnTimer <= 0 {
		e.SpawnTimer = e.SpawnRate

		// Random X position (avoid edges)
		margin := 8

		if e.CodeGrid.Width > (margin * 2) {
			playableWidth := e.CodeGrid.Width - (margin * 2)
			spawnX := rand.Intn(playableWidth) + margin
			e.Enemies = append(e.Enemies, entity.NewBug(spawnX, 0))
		}
	}

	cursorX := e.CodeGrid.CursorX
	cursorY := e.CodeGrid.CursorY
	activeEnemies := e.Enemies[:0]

	for _, bug := range e.Enemies {
		bug.Update(dt)

		// A. HIT DETECTION (Player Kills Bug)
		if bug.IsHit(cursorX, cursorY) {
			e.Explode(bug.X, bug.Y)
			e.Score += 10 // Reward

			// Difficulty Ramp: Every 500 points, spawn 0.1s faster
			if e.Score%100 == 0 && e.SpawnRate > 0.5 {
				e.SpawnRate -= 0.1
			}
			continue
		}

		// Bug Hits Bottom
		if bug.Y >= float64(e.CodeGrid.Height) {
			e.Health -= 10
			// might be nice to add a red flash or something
			if e.Health <= 0 {
				e.IsGameOver = true
			}
			continue
		}

		activeEnemies = append(activeEnemies, bug)
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

	e.drawHUD()

	if e.IsGameOver {
		e.drawGameOver()
	}

	return e.RenderGrid
}

func (e *Engine) drawHUD() {
	status := fmt.Sprintf(" SCORE: %04d | HP: %d%% ", e.Score, e.Health)

	// Color logic: Green if healthy, Red if dying
	hpColor := tcell.ColorGreen
	if e.Health < 50 {
		hpColor = tcell.ColorYellow
	}
	if e.Health < 20 {
		hpColor = tcell.ColorRed
	}

	style := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(hpColor).Bold(true)

	// Draw at Top Right
	startX := e.RenderGrid.Width - len(status) - 2
	y := 0

	for i, r := range status {
		e.RenderGrid.SetContent(startX+i, y, r, style)
	}
}

func (e *Engine) drawGameOver() {
	msg := " SYSTEM FAILURE - PRESS CTRL+\\ TO QUIT "
	style := tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorWhite).Bold(true)

	centerX := (e.RenderGrid.Width - len(msg)) / 2
	centerY := e.RenderGrid.Height / 2

	for i, r := range msg {
		e.RenderGrid.SetContent(centerX+i, centerY, r, style)
	}
}
