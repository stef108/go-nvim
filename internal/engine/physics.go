package engine

import (
	"math"
	"math/rand"

	"neovim-game/internal/core/entity"
)

func (e *Engine) Update(dt float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.IsGameOver {
		return
	}

	e.TimeTotal += dt
	e.updateHeat(dt)
	e.updateCursor(dt)
	e.calculateImpactForce()
	e.updateSpawner(dt)
	e.updateEntities(dt)
}

func (e *Engine) updateHeat(dt float64) {
	//  Decay
	if e.Heat > 0 {
		e.Heat -= e.HeatDecay * dt
	}
	if e.Heat < 0 {
		e.Heat = 0
	}

	//  Check State
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

func (e *Engine) calculateImpactForce() {
	currX := float64(e.CodeGrid.CursorX)
	currY := float64(e.CodeGrid.CursorY)

	// Calculate Distance moved since last frame
	dx := currX - e.LastCursorX
	dy := currY - e.LastCursorY
	dist := math.Sqrt(dx*dx + dy*dy)

	// Set Force (Minimum 1 so you can still touch things)
	e.ImpactForce = max(int(dist), 1)

	e.LastCursorX = currX
	e.LastCursorY = currY
}

func (e *Engine) updateSpawner(dt float64) {
	e.SpawnTimer -= dt
	if e.SpawnTimer <= 0 {
		e.SpawnTimer = e.SpawnRate

		safeW := e.CodeGrid.Width - 5 // So enemies dont spawn on linenumbers
		if safeW > 0 {
			spawnX := rand.Intn(safeW) + 1
			isArmored := rand.Float64() < 0.20
			e.Enemies = append(e.Enemies, entity.NewBug(spawnX, 0, isArmored))
		}
	}
}

func (e *Engine) Explode(x, y float64, force int) {
	e.Particles = append(e.Particles, entity.NewExplosion(x, y, force)...)
}

func (e *Engine) updateEntities(dt float64) {
	cursorHitX := int(e.VisualX + 0.5)
	cursorHitY := int(e.VisualY + 0.5)

	// Process Enemies
	activeEnemies := e.Enemies[:0] // Zero-alloc filtering
	for _, bug := range e.Enemies {
		bug.Update(dt)

		// Collision (Handles both Normal & Railgun)
		if e.checkHit(bug, cursorHitX, cursorHitY) {
			if e.resolveCombat(bug) {
				continue
			}
		}

		// Check Leak
		if bug.Y >= float64(e.CodeGrid.Height) {
			e.handleLeak()
			continue
		}

		// Keep Bug
		activeEnemies = append(activeEnemies, bug)
	}
	e.Enemies = activeEnemies

	// Process Particles
	e.updateParticles(dt)
}

// checkHit determines if a bug was hit based on current mode
func (e *Engine) checkHit(bug *entity.Enemy, cursorX, cursorY int) bool {
	bugX := int(bug.X + 0.5)
	bugY := int(bug.Y)

	if e.IsOverheated {
		// RAILGUN MODE:
		// Hitbox is wider (3 columns) AND infinite height upwards
		hitColumn := bugX >= cursorX-1 && bugX <= cursorX+1
		hitHeight := bugY <= cursorY
		return hitColumn && hitHeight
	}

	// NORMAL MODE:
	// Precise 1x1 hit
	return bug.IsHit(cursorX, cursorY)
}

// Resolvecombat handles the hp / damage logic
func (e *Engine) resolveCombat(bug *entity.Enemy) bool {
	damage := e.ImpactForce

	// Railgun ignores armor and deals infinite damage
	if e.IsOverheated {
		damage = 9999
	}

	// Armor Mechanic: Deflects weak hits
	if bug.IsArmored && damage < 3 {
		// Could spawn a ping partice or sound here later mayber
		return false
	}

	// Apply Damage
	bug.HP -= damage

	if bug.HP <= 0 {
		e.handleKill(bug)
		return true
	}

	return false
}

// handleKill manages Score, Heat, and Explosions
func (e *Engine) handleKill(bug *entity.Enemy) {
	e.Score += 10

	// Big Damage Bonus
	if e.ImpactForce > 20 {
		e.Score += 40 // "CRIT"? Bonus
	}

	if e.IsOverheated {
		e.Explode(bug.X, bug.Y, 50)
		e.Explode(bug.X, bug.Y-1, 50)
	} else {
		// Standard explosion, add Heat to build combo
		e.Explode(bug.X, bug.Y, e.ImpactForce)
		e.Heat += 20.0
	}

	// Difficulty Ramp (Every 500 points)
	if e.Score > 0 && e.Score%500 == 0 && e.SpawnRate > 0.5 {
		e.SpawnRate -= 0.1
	}
}

// handleLeak manages Player Damage
func (e *Engine) handleLeak() {
	e.Health -= 10
	if e.Health <= 0 {
		e.IsGameOver = true
	}
}

// updateParticles manages debris lifecycle
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
