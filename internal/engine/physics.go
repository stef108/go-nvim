package engine

import (
	"math/rand"

	"neovim-game/internal/core/entity"
)

func (e *Engine) Update(dt float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.IsGameOver {
		return
	}

	e.updateCursor(dt)
	e.updateSpawner(dt)
	e.updateEntities(dt)
}

func (e *Engine) updateCursor(dt float64) {
	targetX := float64(e.CodeGrid.CursorX)
	targetY := float64(e.CodeGrid.CursorY)

	e.VisualX += (targetX - e.VisualX) * 25.0 * dt
	e.VisualY += (targetY - e.VisualY) * 25.0 * dt
}

func (e *Engine) updateSpawner(dt float64) {
	e.SpawnTimer -= dt
	if e.SpawnTimer <= 0 {
		e.SpawnTimer = e.SpawnRate

		safeW := e.CodeGrid.Width - 5 // So enemies dont spawn on linenumbers
		if safeW > 0 {
			spawnX := rand.Intn(safeW) + 1
			e.Enemies = append(e.Enemies, entity.NewBug(spawnX, 0))
		}
	}
}

func (e *Engine) updateEntities(dt float64) {
	// Update Enemies
	cursorHitX := int(e.VisualX + 0.5)
	cursorHitY := int(e.VisualY + 0.5)

	activeEnemies := e.Enemies[:0] // Zero-alloc filtering
	for _, bug := range e.Enemies {
		bug.Update(dt)

		// Collision
		if bug.IsHit(cursorHitX, cursorHitY) {
			e.Explode(bug.X, bug.Y)
			e.Score += 10
			// Difficulty Ramp
			if e.Score%100 == 0 && e.SpawnRate > 0.5 {
				e.SpawnRate -= 0.1
			}
			continue
		}

		// Leak
		if bug.Y >= float64(e.CodeGrid.Height) {
			e.Health -= 10
			if e.Health <= 0 {
				e.IsGameOver = true
			}
			continue
		}
		activeEnemies = append(activeEnemies, bug)
	}
	e.Enemies = activeEnemies

	//  Update Particles
	activeParticles := e.Particles[:0]
	for _, p := range e.Particles {
		p.Update(dt)
		if p.Lifetime > 0 && p.Y < float64(e.CodeGrid.Height) && p.Y >= 0 {
			activeParticles = append(activeParticles, p)
		}
	}
	e.Particles = activeParticles
}

func (e *Engine) Explode(x, y float64) {
	e.Particles = append(e.Particles, entity.NewExplosion(x, y)...)
}
