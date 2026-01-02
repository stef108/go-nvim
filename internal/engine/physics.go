package engine

import (
	"math"
	"math/rand"

	"neovim-game/internal/core/entity"
)

// Update is the Master Loop
func (e *Engine) Update(dt float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.IsGameOver {
		return
	}

	e.updateHeat(dt)
	e.updateCursor(dt)
	e.calculateImpactForce()

	e.updateSpawner(dt)
	e.updateEntities(dt)
}

// updateEntities manages the lifecycle of all game objects
func (e *Engine) updateEntities(dt float64) {
	ctx := entity.Context{
		DT:      dt,
		Grid:    e.CodeGrid,
		PlayerX: int(e.VisualX + 0.5),
		PlayerY: int(e.VisualY + 0.5),
	}

	activeEntities := e.Entities[:0]

	for _, ent := range e.Entities {
		// A. Run Entity AI/Physics
		ent.Update(ctx)

		// B. Interaction (Combat, Collision)
		if e.processInteraction(ent, ctx) {
			continue
		}

		// C. Boundaries (Did it fall off screen?)
		if e.processLeak(ent) {
			continue
		}

		// D. Keep it alive
		activeEntities = append(activeEntities, ent)
	}
	e.Entities = activeEntities

	// Particles (Separate loop as they don't interact)
	e.updateParticles(dt)
}

// processInteraction handles ALL Player vs Entity logic
func (e *Engine) processInteraction(ent entity.GameEntity, ctx entity.Context) bool {
	hit := false

	if e.IsOverheated {
		ex, ey := ent.Position()
		hit = int(ex) >= ctx.PlayerX-1 && int(ex) <= ctx.PlayerX+1 && int(ey) <= ctx.PlayerY
	} else {
		hit = ent.IsHit(ctx.PlayerX, ctx.PlayerY)
	}

	if hit {
		damage := e.ImpactForce
		if e.IsOverheated {
			damage = 9999
		}

		// Let the entity decide if it dies (Armor/HP logic)
		dead := ent.TakeDamage(damage)

		if dead {
			ex, ey := ent.Position()
			e.handleKill(ex, ey)
			return true
		}
	}

	// If Cursor is literally inside the entity
	if ent.IsHit(e.CodeGrid.CursorX, e.CodeGrid.CursorY) {
		if rand.Float64() < 0.1 {
			e.handlePlayerDamage(10)
			ex, ey := ent.Position()
			e.Explode(ex, ey, 1)
		}
	}

	return false // Entity survives
}

// processLeak handles entities hitting the bottom of the screen
func (e *Engine) processLeak(ent entity.GameEntity) bool {
	_, y := ent.Position()
	if y >= float64(e.CodeGrid.Height) {
		e.handlePlayerDamage(10)
		return true // Remove entity
	}
	return false
}

// --- HELPERS ---

func (e *Engine) handleKill(x, y float64) {
	e.Score += 10
	if e.ImpactForce > 20 {
		e.Score += 40
	}

	// Visuals
	if e.IsOverheated {
		e.Explode(x, y, 50)
		e.Explode(x, y-1, 50)
	} else {
		e.Explode(x, y, e.ImpactForce)
		e.Heat += 20.0
	}

	// Difficulty Ramp
	if e.Score > 0 && e.Score%500 == 0 && e.SpawnRate > 0.5 {
		e.SpawnRate -= 0.1
	}
}

func (e *Engine) handlePlayerDamage(amount int) {
	e.Health -= amount
	if e.Health <= 0 {
		e.IsGameOver = true
	}
}

func (e *Engine) calculateImpactForce() {
	currX := float64(e.CodeGrid.CursorX)
	currY := float64(e.CodeGrid.CursorY)

	dx := currX - e.LastCursorX
	dy := currY - e.LastCursorY
	dist := math.Sqrt(dx*dx + dy*dy)

	e.ImpactForce = int(dist)
	if e.ImpactForce < 1 {
		e.ImpactForce = 1
	}

	e.LastCursorX = currX
	e.LastCursorY = currY
}

func (e *Engine) updateHeat(dt float64) {
	if e.Heat > 0 {
		e.Heat -= e.HeatDecay * dt
	}
	if e.Heat < 0 {
		e.Heat = 0
	}

	if e.Heat >= e.MaxHeat {
		e.IsOverheated = true
		e.Heat = e.MaxHeat
	} else if e.IsOverheated && e.Heat <= 0 {
		e.IsOverheated = false
	}
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

		// 30% Crawler, 70% Bug
		if rand.Float64() < 0.30 {
			e.spawnCrawler()
		} else {
			e.spawnBug()
		}
	}
}

func (e *Engine) spawnBug() {
	safeW := e.CodeGrid.Width - 5
	if safeW > 0 {
		spawnX := rand.Intn(safeW) + 1
		isArmored := rand.Float64() < 0.20
		e.Entities = append(e.Entities, entity.NewBug(spawnX, 0, isArmored))
	}
}

func (e *Engine) spawnCrawler() {
	// Try to find a perch
	for i := 0; i < 10; i++ {
		rx := rand.Intn(e.CodeGrid.Width)
		ry := rand.Intn(e.CodeGrid.Height / 2)
		if e.CodeGrid.IsSolid(rx, ry) && !e.CodeGrid.IsSolid(rx, ry-1) {
			e.Entities = append(e.Entities, entity.NewCrawler(rx, ry-1))
			return
		}
	}
}

func (e *Engine) Explode(x, y float64, force int) {
	e.Particles = append(e.Particles, entity.NewExplosion(x, y, force)...)
}

func (e *Engine) updateParticles(dt float64) {
	activeParticles := e.Particles[:0]
	for _, p := range e.Particles {
		p.Update(dt)
		if p.Lifetime > 0 && p.Y < float64(e.CodeGrid.Height) && p.Y >= 0 {
			activeParticles = append(activeParticles, p)
		}
	}
	e.Particles = activeParticles
}
