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

	// Visuals
	VisualX float64
	VisualY float64

	// System
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
