package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/neovim/go-client/nvim"
)

func main() {
	fmt.Println("--- DIAGNOSTIC TOOL START ---")

	// 1. Start Neovim
	cmd := exec.Command("nvim", "--embed", "--clean")
	v, err := nvim.NewChildProcess(
		nvim.ChildProcessCommand(cmd.Path),
		nvim.ChildProcessArgs(cmd.Args[1:]...),
		nvim.ChildProcessLogf(log.Printf),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer v.Close()
	fmt.Println("1. Neovim Process Started")

	// 2. Register Handler (The part that seems broken)
	// We use a generic interface{} to catch EVERYTHING
	v.RegisterHandler("redraw", func(updates []interface{}) {
		fmt.Printf("!!! RECEIVED REDRAW EVENT (Batch of %d updates) !!!\n", len(updates))
		for i, update := range updates {
			if i < 3 { // Print first 3 events only
				fmt.Printf("   - Event: %v\n", update)
			}
		}
	})
	fmt.Println("2. Handler Registered")

	// 3. Attach UI
	opts := map[string]any{"ext_linegrid": true, "rgb": true}
	if err := v.AttachUI(80, 24, opts); err != nil {
		log.Fatal(err)
	}
	fmt.Println("3. UI Attached (Neovim should send events NOW)")

	// 4. Send Test Input
	fmt.Println("4. Sending keystrokes...")
	v.Input("i")
	v.Input("h")
	v.Input("e")
	v.Input("l")
	v.Input("l")
	v.Input("o")
	v.Input("<Esc>")

	// 5. Check RPC Liveness
	out, err := v.CommandOutput("echo 'RPC_IS_ALIVE'")
	if err != nil {
		fmt.Printf("RPC Command Failed: %v\n", err)
	} else {
		fmt.Printf("5. RPC Response: %s\n", out)
	}

	// Wait 2 seconds to catch any delayed logs
	time.Sleep(2 * time.Second)
	fmt.Println("--- DIAGNOSTIC FINISHED ---")
}
