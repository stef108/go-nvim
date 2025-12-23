package main

import (
	"log"
	"time"

	"neovim-game/internal/adapters/nvim"
	"neovim-game/internal/adapters/tui"
	"neovim-game/internal/common"
	"neovim-game/internal/engine"

	"github.com/gdamore/tcell/v2"
)

func main() {
	f := common.InitLogger()
	defer f.Close()
	log.Println("--- STARTING GAME ENGINE ---")

	// 1. Init TUI
	renderer, err := tui.New()
	if err != nil {
		log.Fatal(err)
	}
	defer renderer.Close()
	w, h := renderer.Size()

	// 2. Init Engine (The Brain)
	game := engine.New(w, h)

	// 3. Init Neovim (The Background)
	nv, err := nvim.New(w, h, game.CodeGrid)
	if err != nil {
		renderer.Close()
		log.Fatalf("Neovim failed: %v", err)
	}
	defer nv.Close()

	// 4. Input Channels
	inputChan := make(chan string)
	resizeChan := make(chan struct{ w, h int })

	// 5. Start Input Loop
	go func() {
		for {
			ev := renderer.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventResize:
				width, height := ev.Size()
				renderer.Sync()
				resizeChan <- struct{ w, h int }{width, height}
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyCtrlBackslash {
					renderer.Close()
					return
				}
				// Game Logic: Spawn bug on 's' key for testing
				if ev.Rune() == 's' {
					// We spawn it at column 10, row 0
					log.Println("Spawning Bug!")
					game.SpawnEnemy(10, 0)
				}

				vimKey := tui.InputToVimString(ev)
				if vimKey != "" {
					select {
					case inputChan <- vimKey:
					default:
					}
				}
			}
		}
	}()

	// 6. Handle Channels
	go func() {
		for {
			select {
			case key := <-inputChan:
				nv.Input(key)
			case size := <-resizeChan:
				game.Resize(size.w, size.h)
				nv.Resize(size.w, size.h)
			}
		}
	}()
	const targetFPS = 120
	ticker := time.NewTicker(time.Second / time.Duration(targetFPS))
	defer ticker.Stop()

	lastTime := time.Now()

	for range ticker.C {
		now := time.Now()
		dt := now.Sub(lastTime).Seconds()
		lastTime = now

		// A. Update Physics
		game.Update(dt)

		// B. Compose Frame (Code + Enemies)
		finalGrid := game.DrawCompositor()

		// C. Render to Screen
		renderer.Render(finalGrid)
	}
}
