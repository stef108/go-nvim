package engine

import (
	"fmt"

	"neovim-game/internal/core/grid"

	"github.com/gdamore/tcell/v2"
)

func (e *Engine) DrawCompositor() *grid.Grid {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Background
	e.RenderGrid.CopyFrom(e.CodeGrid)

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
