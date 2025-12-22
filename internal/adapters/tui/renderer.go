package tui

import (
	"fmt"

	"neovim-game/internal/core/grid"

	"github.com/gdamore/tcell/v2"
)

// Renderer handles the physical terminal screen.
type Renderer struct {
	Screen tcell.Screen
	width  int
	height int
}

// New initializes the tcell screen.
func New() (*Renderer, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := s.Init(); err != nil {
		return nil, err
	}

	// Enable mouse support if you want users to be able to click (optional)
	s.EnableMouse()

	// Set default style
	s.SetStyle(tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset))

	w, h := s.Size()

	return &Renderer{
		Screen: s,
		width:  w,
		height: h,
	}, nil
}

// Close shuts down the screen and restores the terminal.
func (r *Renderer) Close() {
	r.Screen.Fini()
}

// Size returns the current dimensions.
func (r *Renderer) Size() (int, int) {
	return r.Screen.Size()
}

// Sync updates our internal size trackers. Call this on resize events.
func (r *Renderer) Sync() {
	r.width, r.height = r.Screen.Size()
	r.Screen.Sync()
}

// Render commits a Grid to the physical screen.
// This is the Frame Draw call.
func (r *Renderer) Render(g *grid.Grid) {
	// We lock the grid to ensure we don't read while Neovim is writing

	r.Screen.ShowCursor(g.CursorX, g.CursorY)

	// Optimize: Only draw within the bounds of the screen or grid, whichever is smaller
	renderW := r.width
	if g.Width < renderW {
		renderW = g.Width
	}
	renderH := r.height
	if g.Height < renderH {
		renderH = g.Height
	}

	for y := 0; y < renderH; y++ {
		for x := 0; x < renderW; x++ {
			char, style, width := g.GetContent(x, y)

			// Tcell optimization: smart invalidation is built-in to SetContent,
			// but we can skip 'width 0' cells (continuations) as SetContent handles them via the previous call.
			if width == 0 {
				continue
			}

			// Draw the cell
			// combinedArgs is for combining characters (accents), we usually don't need them for code.
			r.Screen.SetContent(x, y, char, nil, style)
		}
	}

	// The actual flush to the terminal
	r.Screen.Show()
}

// PollEvent wraps the tcell event polling.
// This is blocking, so run it in a loop.
func (r *Renderer) PollEvent() tcell.Event {
	return r.Screen.PollEvent()
}

// InputToVimString converts tcell events to Vim notation.
// e.g. Ctrl+C -> "<C-c>", Escape -> "<Esc>"
func InputToVimString(ev *tcell.EventKey) string {
	// 1. Handle Special Keys First
	switch ev.Key() {
	case tcell.KeyEsc:
		return "<Esc>"
	case tcell.KeyEnter:
		return "<CR>"
	case tcell.KeyTab:
		return "<Tab>"
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		return "<BS>"
	case tcell.KeyDelete:
		return "<Del>"
	case tcell.KeyUp:
		return "<Up>"
	case tcell.KeyDown:
		return "<Down>"
	case tcell.KeyLeft:
		return "<Left>"
	case tcell.KeyRight:
		return "<Right>"
	case tcell.KeyCtrlSpace:
		return " "
	}

	// 2. Handle Control Modifiers (Ctrl+A, Ctrl+C, etc)
	if ev.Modifiers()&tcell.ModCtrl != 0 {
		if ev.Key() == tcell.KeyCtrlC {
			return "<C-c>"
		}

		// Generic Ctrl+Letter
		if ev.Key() >= tcell.KeyCtrlA && ev.Key() <= tcell.KeyCtrlZ {
			char := rune('a' + (ev.Key() - tcell.KeyCtrlA))
			return fmt.Sprintf("<C-%c>", char)
		}
	}

	// Standard Runes (Letters, Numbers, Symbols)
	if ev.Rune() != 0 {
		return string(ev.Rune())
	}

	return ""
}
