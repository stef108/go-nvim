package engine

import (
	"sync"

	"neovim-game/internal/core/entity"
	"neovim-game/internal/core/grid"
)

type Engine struct {
	// State
	Entities  []entity.GameEntity
	Particles []*entity.Particle

	// The Buffers
	CodeGrid   *grid.Grid // The Clean state from Neovim
	RenderGrid *grid.Grid // The altered state sent to TUI

	// Visuals
	VisualX float64
	VisualY float64

	// System
	stopChan  chan struct{}
	mu        sync.RWMutex
	TimeTotal float64

	// Game Rules
	Score      int
	Health     int
	MaxHealth  int
	IsGameOver bool

	// impact system
	LastCursorX float64
	LastCursorY float64
	ImpactForce int

	// Combo System
	Heat         float64
	MaxHeat      float64
	HeatDecay    float64
	IsOverheated bool

	// Wave Logic
	SpawnTimer float64
	SpawnRate  float64
}

func New(w, h int) *Engine {
	return &Engine{
		Entities:   make([]entity.GameEntity, 0),
		Particles:  make([]*entity.Particle, 0),
		CodeGrid:   grid.New(w, h),
		RenderGrid: grid.New(w, h),
		stopChan:   make(chan struct{}),

		// Defaults
		Score:        0,
		Health:       100,
		MaxHealth:    100,
		SpawnRate:    2.0,
		SpawnTimer:   0,
		Heat:         0,
		MaxHeat:      100,
		HeatDecay:    2.5,
		IsOverheated: false,
	}
}
