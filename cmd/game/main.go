package main

import (
	"log"

	"neovim-game/internal/common"
)

func main() {
	f := common.InitLogger()
	defer f.Close()
	log.Println("--- STARTING NEOVIM GAME ---")

	// Run
	app := NewApp(120)
	app.Run()
}
