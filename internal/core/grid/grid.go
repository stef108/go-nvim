package grid

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// Cell represents a single terminal cell.
type Cell struct {
	Char  rune
	Style tcell.Style
	Width int
}

// Grid is a 2D buffer of cells.
type Grid struct {
	Width   int
	Height  int
	Cells   []Cell
	mu      sync.RWMutex // Protects resizing operations
	CursorX int
	CursorY int
}

// New creates a grid of specific dimensions.
func New(width, height int) *Grid {
	return &Grid{
		Width:  width,
		Height: height,
		Cells:  make([]Cell, width*height),
	}
}

func (g *Grid) SetCursor(x, y int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.CursorX = x
	g.CursorY = y
}

// Resize safely resizes the grid, preserving old content where it overlaps.
func (g *Grid) Resize(w, h int) {
	g.mu.Lock()
	defer g.mu.Unlock()

	newCells := make([]Cell, w*h)

	// Copy logic: iterate over the *smaller* of the old/new dimensions
	minH := h
	if g.Height < h {
		minH = g.Height
	}
	minW := w
	if g.Width < w {
		minW = g.Width
	}

	for y := 0; y < minH; y++ {
		for x := 0; x < minW; x++ {
			// Map 2D (x,y) to 1D index
			oldIdx := y*g.Width + x
			newIdx := y*w + x
			newCells[newIdx] = g.Cells[oldIdx]
		}
	}

	g.Cells = newCells
	g.Width = w
	g.Height = h
}

// SetContent writes a character to the grid.
func (g *Grid) SetContent(x, y int, mainc rune, style tcell.Style) {
	if x < 0 || y < 0 || x >= g.Width || y >= g.Height {
		return // Bounds check
	}

	idx := y*g.Width + x
	// Calculate width
	w := runewidth.RuneWidth(mainc)

	// Write the cell
	g.Cells[idx] = Cell{
		Char:  mainc,
		Style: style,
		Width: w,
	}

	// If it's a wide character (width 2), we must mark the NEXT cell
	// as a placeholder so we don't draw over it.
	if w == 2 && x+1 < g.Width {
		g.Cells[idx+1] = Cell{
			Char:  ' ', // Placeholder
			Style: style,
			Width: 0, // 0 indicates "continuation of previous"
		}
	}
}

// GetContent retrieves content for rendering.
func (g *Grid) GetContent(x, y int) (rune, tcell.Style, int) {
	if x < 0 || y < 0 || x >= g.Width || y >= g.Height {
		return ' ', tcell.StyleDefault, 1
	}
	c := g.Cells[y*g.Width+x]
	return c.Char, c.Style, c.Width
}

// Clear wipes the grid efficiently.
func (g *Grid) Clear() {
	// Fast memory clear usually works, but for Tcell we need correct defaults
	empty := Cell{Char: ' ', Style: tcell.StyleDefault, Width: 1}
	for i := range g.Cells {
		g.Cells[i] = empty
	}
}

// CopyFrom overlays one grid onto another.
// Useful for stamping the "Game Entity Layer" onto the "Background Layer".
func (dest *Grid) Overlay(src *Grid, offsetX, offsetY int) {
	dest.mu.Lock() // Write lock on destination
	defer dest.mu.Unlock()

	// We don't lock source (assume it's a read-only snapshot or owned by thread)

	for y := 0; y < src.Height; y++ {
		destY := y + offsetY
		if destY >= dest.Height {
			break
		}

		for x := 0; x < src.Width; x++ {
			destX := x + offsetX
			if destX >= dest.Width {
				break
			}

			idx := y*src.Width + x
			cell := src.Cells[idx]

			// For now, we copy everything.
			if cell.Width > 0 {
				destIdx := destY*dest.Width + destX
				dest.Cells[destIdx] = cell

				// Handle wide chars in copy
				if cell.Width == 2 && destX+1 < dest.Width {
					dest.Cells[destIdx+1] = Cell{Width: 0}
				}
			}
		}
	}
}

func (g *Grid) CopyFrom(src *Grid) {
	g.mu.Lock()
	defer g.mu.Unlock()

	src.mu.RLock()
	defer src.mu.RUnlock()

	// ... copy loop ...
	copy(g.Cells, src.Cells)

	// Copy dimensions
	g.Width = src.Width
	g.Height = src.Height

	// Copy Cursor
	g.CursorX = src.CursorX
	g.CursorY = src.CursorY
}
