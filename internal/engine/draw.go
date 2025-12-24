package engine

import (
	"fmt"
	"math"
	"math/rand"

	"neovim-game/internal/core/grid"

	"github.com/gdamore/tcell/v2"
)

func (e *Engine) DrawCompositor() *grid.Grid {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Background
	e.RenderGrid.CopyFrom(e.CodeGrid)

	if e.IsOverheated {
		e.drawRailgun()
	}
	// Draw Entities
	for _, bug := range e.Enemies {
		x, y := bug.RenderPos()
		if e.isInBounds(x, y) {
			e.RenderGrid.SetContent(x, y, bug.Char, bug.Style)
		}
	}

	for _, p := range e.Particles {
		x, y := int(p.X), int(p.Y)
		if e.isInBounds(x, y) {
			e.RenderGrid.SetContent(x, y, p.Char, p.Style)
		}
	}

	// Draw UI
	e.drawHUD()
	e.drawHeatGauge()

	if e.IsGameOver {
		e.drawGameOver()
	}

	// Sync Visual Cursor
	e.RenderGrid.CursorX = int(e.VisualX + 0.5)
	e.RenderGrid.CursorY = int(e.VisualY + 0.5)

	return e.RenderGrid
}

// Inline helper for bounds checking
func (e *Engine) isInBounds(x, y int) bool {
	return x >= 0 && x < e.RenderGrid.Width && y >= 0 && y < e.RenderGrid.Height
}

func (e *Engine) drawHUD() {
	status := fmt.Sprintf(" SCORE: %04d | HP: %d%% ", e.Score, e.Health)
	hpColor := tcell.ColorGreen
	if e.Health < 50 {
		hpColor = tcell.ColorYellow
	}
	if e.Health < 20 {
		hpColor = tcell.ColorRed
	}

	style := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(hpColor).Bold(true)

	startX := e.RenderGrid.Width - len(status) - 2
	for i, r := range status {
		e.RenderGrid.SetContent(startX+i, 0, r, style)
	}
}

func (e *Engine) drawGameOver() {
	msg := " SYSTEM FAILURE - PRESS CTRL+\\ TO QUIT "
	style := tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorWhite).Bold(true)
	cx := (e.RenderGrid.Width - len(msg)) / 2
	cy := e.RenderGrid.Height / 2
	for i, r := range msg {
		e.RenderGrid.SetContent(cx+i, cy, r, style)
	}
}

func (e *Engine) drawRailgun() {
	cx := int(e.VisualX + 0.5)
	cy := int(e.VisualY + 0.5)

	// Animation speed multipliers
	scrollSpeed := e.TimeTotal * 15.0
	pulseSpeed := e.TimeTotal * 8.0

	// Define the color palette
	cyan := tcell.ColorAqua
	white := tcell.ColorWhite
	// Oscillate between Cyan and White based on time
	colorLerp := math.Sin(pulseSpeed) // Goes from -1 to 1

	coreColor := cyan
	if colorLerp > 0.5 {
		coreColor = white // Flash white periodically
	}

	// Styles
	coreStyle := tcell.StyleDefault.Foreground(coreColor).Background(cyan).Bold(true)
	glowStyle := tcell.StyleDefault.Foreground(cyan).Background(tcell.ColorReset)

	// The characters for the scrolling core
	coreChars := []rune{'▒', '▓', '█', '▓'}

	for y := cy; y >= 0; y-- {
		// Use Y position + Time to determine which character to draw.
		charIdx := int(float64(y)/2.0-scrollSpeed) % len(coreChars)
		if charIdx < 0 {
			charIdx += len(coreChars)
		}

		coreChar := coreChars[charIdx]
		e.RenderGrid.SetContent(cx, y, coreChar, coreStyle)

		// Use random noise to determine width of the glow this frame.
		noise := rand.Float64()

		// Always draw inner glow
		e.RenderGrid.SetContent(cx-1, y, '░', glowStyle)
		e.RenderGrid.SetContent(cx+1, y, '░', glowStyle)

		// 30% chance to draw wider glow (the crackle)
		if noise > 0.7 {
			e.RenderGrid.SetContent(cx-2, y, '·', glowStyle)
			e.RenderGrid.SetContent(cx+2, y, '·', glowStyle)
		}
	}
}

func (e *Engine) drawHeatGauge() {
	// Draw a bar at the bottom center
	barWidth := 40
	startX := (e.RenderGrid.Width - barWidth) / 2
	y := e.RenderGrid.Height - 1

	// Calculate fill amount
	fillPercent := e.Heat / e.MaxHeat
	if fillPercent > 1.0 {
		fillPercent = 1.0
	}
	filledChars := int(float64(barWidth) * fillPercent)

	// Draw the Container
	for i := 0; i < barWidth; i++ {
		char := ' '
		style := tcell.StyleDefault.Background(tcell.ColorGray)

		if i < filledChars {
			// Simple threshold approach:
			if e.IsOverheated {
				style = style.Background(tcell.ColorAqua).Foreground(tcell.ColorWhite)
				char = '⚡'
			} else if fillPercent > 0.8 {
				style = style.Background(tcell.ColorRed)
				char = '▓'
			} else if fillPercent > 0.5 {
				style = style.Background(tcell.ColorOrange)
				char = '▒'
			} else {
				style = style.Background(tcell.ColorGreen)
				char = '░'
			}
		}

		e.RenderGrid.SetContent(startX+i, y, char, style)
	}

	// Label
	label := " HEAT "
	if e.IsOverheated {
		label = " OVERDRIVE "
	}
	labelX := startX + (barWidth-len(label))/2
	labelStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)

	for i, r := range label {
		e.RenderGrid.SetContent(labelX+i, y, r, labelStyle)
	}
}
