package main

import (
	"log"
	"time"

	"neovim-game/internal/adapters/nvim"
	"neovim-game/internal/adapters/tui"
	"neovim-game/internal/common"
	"neovim-game/internal/core/grid"

	"github.com/gdamore/tcell/v2"
)

func main() {
	f := common.InitLogger()
	defer f.Close()
	log.Println("--- STARTING NEOVIM GAME ---")

	// Init Render Screen
	renderer, err := tui.New()
	if err != nil {
		log.Fatalf("Fatal: Tcell init failed: %v", err)
	}
	defer renderer.Close()

	w, h := renderer.Size()
	mainGrid := grid.New(w, h)

	// --- TEST PATTERN (Draw Red Diagonal) ---
	for i := 0; i < 20 && i < w && i < h; i++ {
		mainGrid.SetContent(i, i, '█', tcell.StyleDefault.Foreground(tcell.ColorRed))
	}
	renderer.Render(mainGrid)
	log.Println("--- TEST PATTERN DRAWN ---")
	// ----------------------------------------

	// Start Input Loop (NON-BLOCKING)
	// We start this NOW so we can capture Ctrl+C even if Neovim hangs
	inputChan := make(chan string)              // Channel to send keys to Neovim
	resizeChan := make(chan struct{ w, h int }) // Channel for resize events

	go func() {
		log.Println("--- INPUT LOOP STARTED ---")
		for {
			ev := renderer.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventResize:
				width, height := ev.Size()
				renderer.Sync()
				resizeChan <- struct{ w, h int }{width, height}

			case *tcell.EventKey:
				log.Printf("Key: %v", ev.Name()) // DEBUG LOG

				if ev.Key() == tcell.KeyCtrlBackslash {
					log.Println("Emergency Exit Triggered")
					renderer.Close()
					return // Hard exit
				}

				vimKey := tui.InputToVimString(ev)
				if vimKey != "" {
					// Non-blocking send (so TUI doesn't freeze if Neovim is busy)
					select {
					case inputChan <- vimKey:
					default:
					}
				}
			}
		}
	}()

	// Init Neovim (Blocking, but with timeout)
	log.Println("--- CONNECTING TO NEOVIM ---")
	nv, err := nvim.New(w, h, mainGrid)
	if err != nil {
		renderer.Close()
		log.Fatalf("Neovim failed: %v", err)
	}
	defer nv.Close()

	// Wire up Channels
	// Listen for input from our TUI loop and send to Neovim
	go func() {
		for {
			select {
			case key := <-inputChan:
				nv.Input(key)
			case size := <-resizeChan:
				mainGrid.Resize(size.w, size.h)
				nv.Resize(size.w, size.h)
			}
		}
	}()

	// The Render Trigger
	nv.OnFlush = func() {
		mainGrid.CursorX = nv.CursorX
		mainGrid.CursorY = nv.CursorY
		renderer.Render(mainGrid)
	}

	// Force Neovim to Draw Initial Screen
	// We send a command to force a redraw, clearing the Red Line
	go func() {
		time.Sleep(500 * time.Millisecond)
		nv.Input("<Esc>")
		nv.Input(":redraw!<CR>") // Force full repaint
		log.Println("--- SENT REDRAW COMMAND ---")
	}()

	// Block forever
	select {}
}
