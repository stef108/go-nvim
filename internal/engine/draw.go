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

func (e *Engine) drawHUD() {
	// We draw on the very last line of the screen
	y := e.RenderGrid.Height - 1
	w := e.RenderGrid.Width

	// 1. Clear the HUD Background (Dark Gray/Black for contrast)
	bgStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	for x := range w {
		e.RenderGrid.SetContent(x, y, ' ', bgStyle)
	}

	// 2. Draw Left Module: Health
	// Layout: [♥ 100% ■■■■■■■■■■]
	e.drawHealthModule(0, y)

	// 3. Draw Right Module: Score
	// Layout: [SCORE 001500]
	e.drawScoreModule(w, y)

	// 4. Draw Center Module: Heat/Overdrive
	// Layout:  --[ HEAT ⚡⚡⚡ ]--
	// We calculate remaining space to center it perfectly
	e.drawHeatModule(w, y)
}

func (e *Engine) drawHealthModule(startX, y int) {
	// Health Color Logic
	hpColor := tcell.ColorGreen
	if e.Health < 50 {
		hpColor = tcell.ColorYellow
	}
	if e.Health < 20 {
		hpColor = tcell.ColorRed
	}

	// The Label " HP "
	labelStyle := tcell.StyleDefault.Background(hpColor).Foreground(tcell.ColorBlack).Bold(true)
	label := " HP "
	cursor := startX

	for _, r := range label {
		e.RenderGrid.SetContent(cursor, y, r, labelStyle)
		cursor++
	}

	// The Bar [■■■□□]
	// We use 10 blocks for 100% health
	blocks := 10
	fill := int((float64(e.Health) / float64(e.MaxHealth)) * float64(blocks))

	// Add a little padding
	cursor++

	for i := range blocks {
		char := '▱'
		style := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(tcell.ColorBlack)

		if i < fill {
			char = '▰'
			style = tcell.StyleDefault.Foreground(hpColor).Background(tcell.ColorBlack)
		}
		e.RenderGrid.SetContent(cursor, y, char, style)
		cursor++
	}

	// Draw numerical percent
	pct := fmt.Sprintf(" %d%%", e.Health)
	for _, r := range pct {
		e.RenderGrid.SetContent(cursor, y, r, tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
		cursor++
	}
}

func (e *Engine) drawScoreModule(screenWidth, y int) {
	// Format: "SCORE 000000"
	text := fmt.Sprintf("SCORE %06d ", e.Score)

	// Color: Neon Blue for arcade feel
	style := tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite).Bold(true)

	// Draw from Right to Left
	startX := screenWidth - len(text)
	for i, r := range text {
		e.RenderGrid.SetContent(startX+i, y, r, style)
	}
}

func (e *Engine) drawHeatModule(screenWidth, y int) {
	// We want this centered between the Health (Left) and Score (Right)

	barWidth := 30
	startX := (screenWidth - barWidth) / 2

	fillPercent := e.Heat / e.MaxHeat
	if fillPercent > 1.0 {
		fillPercent = 1.0
	}

	// If Overheated, the bar pulses or looks different
	baseChar := '═'
	fillChar := '═'
	primaryColor := tcell.ColorOrange

	if e.IsOverheated {
		primaryColor = tcell.ColorAqua
		fillChar = '≡'
	}

	// Draw the "Track"
	centerLabel := " HEAT "
	if e.IsOverheated {
		centerLabel = " OVERDRIVE "
	}

	// Draw the Bar container
	filledLen := int(float64(barWidth) * fillPercent)

	for i := range barWidth {
		char := baseChar
		style := tcell.StyleDefault.Foreground(tcell.ColorDarkGray).Background(tcell.ColorBlack)

		// Fill Logic
		if i < filledLen {
			char = fillChar
			// Gradient Logic
			col := primaryColor
			if !e.IsOverheated {
				if fillPercent > 0.8 {
					col = tcell.ColorRed
				}
				if fillPercent < 0.5 {
					col = tcell.ColorGreen
				}
			}
			style = tcell.StyleDefault.Foreground(col).Background(tcell.ColorBlack)
		}

		e.RenderGrid.SetContent(startX+i, y, char, style)
	}

	// Overlay the Label in the absolute center of the bar
	labelX := startX + (barWidth-len(centerLabel))/2
	labelStyle := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack).Bold(true)

	if e.IsOverheated {
		labelStyle = labelStyle.Background(tcell.ColorAqua).Foreground(tcell.ColorBlack)
	}

	for i, r := range centerLabel {
		e.RenderGrid.SetContent(labelX+i, y, r, labelStyle)
	}
}
