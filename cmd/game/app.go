package main

import (
	"log"
	"time"

	"neovim-game/internal/adapters/nvim"
	"neovim-game/internal/adapters/tui"
	"neovim-game/internal/engine"

	"github.com/gdamore/tcell/v2"
)

type App struct {
	Renderer *tui.Renderer
	Adapter  *nvim.Adapter
	Engine   *engine.Engine

	// Config
	TargetFPS int
}

func NewApp(fps int) *App {
	//  Init TUI
	renderer, err := tui.New()
	if err != nil {
		log.Fatal(err)
	}

	w, h := renderer.Size()

	// Init Engine
	game := engine.New(w, h)

	// Init Neovim
	nv, err := nvim.New(w, h, game.CodeGrid)
	if err != nil {
		renderer.Close()
		log.Fatalf("Neovim failed: %v", err)
	}

	return &App{
		Renderer:  renderer,
		Adapter:   nv,
		Engine:    game,
		TargetFPS: fps,
	}
}

func (app *App) Run() {
	defer app.Renderer.Close()
	defer app.Adapter.Close()

	// Channels
	inputChan := make(chan string, 100) // Buffered slightly
	resizeChan := make(chan struct{ w, h int })

	// --- Input Loop ---
	go func() {
		for {
			ev := app.Renderer.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventResize:
				w, h := ev.Size()
				app.Renderer.Sync()
				resizeChan <- struct{ w, h int }{w, h}
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyCtrlBackslash {
					app.Renderer.Close()
					return
				}
				vimKey := tui.InputToVimString(ev)
				if vimKey != "" {
					select {
					case inputChan <- vimKey:
					default: // Drop key if Neovim is hanging
					}
				}
			}
		}
	}()

	// --- Neovim Bridge Loop ---
	go func() {
		for {
			select {
			case key := <-inputChan:
				app.Adapter.Input(key)
			case size := <-resizeChan:
				app.Engine.Resize(size.w, size.h)
				app.Adapter.Resize(size.w, size.h)
			}
		}
	}()

	// --- The Game Loop ---
	ticker := time.NewTicker(time.Second / time.Duration(app.TargetFPS))
	defer ticker.Stop()

	lastTime := time.Now()

	for range ticker.C {
		now := time.Now()
		dt := now.Sub(lastTime).Seconds()
		lastTime = now

		app.Engine.Update(dt)
		frame := app.Engine.DrawCompositor()
		app.Renderer.Render(frame)
	}
}
