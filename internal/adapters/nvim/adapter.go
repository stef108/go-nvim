package nvim

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"time"

	"neovim-game/internal/core/grid"

	"github.com/gdamore/tcell/v2"
	"github.com/neovim/go-client/nvim"
)

// Helper to safely cast any number type to int
func toInt(v any) int {
	switch i := v.(type) {
	case int64:
		return int(i)
	case uint64:
		return int(i)
	case float64:
		return int(i)
	case int:
		return i
	default:
		return 0
	}
}

type Adapter struct {
	Ref     *nvim.Nvim
	Grid    *grid.Grid
	styles  map[int]tcell.Style
	CursorX int
	CursorY int
	OnFlush func()
}

func New(width, height int, targetGrid *grid.Grid) (*Adapter, error) {
	cmd := exec.Command("nvim", "--embed")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	type result struct {
		adapter *Adapter
		err     error
	}
	done := make(chan result, 1)

	go func() {
		v, err := nvim.NewChildProcess(
			nvim.ChildProcessCommand(cmd.Path),
			nvim.ChildProcessArgs(cmd.Args[1:]...),
			nvim.ChildProcessLogf(log.Printf),
		)
		if err != nil {
			done <- result{nil, fmt.Errorf("start failed: %v", err)}
			return
		}

		adapter := &Adapter{
			Ref:    v,
			Grid:   targetGrid,
			styles: make(map[int]tcell.Style),
		}
		// Use the variadic handler signature
		v.RegisterHandler("redraw", adapter.handleRedraw)

		opts := map[string]any{"ext_linegrid": true, "rgb": true}
		if err := v.AttachUI(width, height, opts); err != nil {
			v.Close()
			done <- result{nil, fmt.Errorf("AttachUI failed: %v | Stderr: %s", err, stderr.String())}
			return
		}
		v.Command("set virtualedit=all")
		v.Command("set scrolloff=0")
		v.Command("set sidescrolloff=0")
		v.Command("set nowrap")
		done <- result{adapter, nil}
	}()

	select {
	case res := <-done:
		return res.adapter, res.err
	case <-time.After(2 * time.Second):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return nil, fmt.Errorf("timeout. Stderr: %s", stderr.String())
	}
}

func (a *Adapter) Close()          { a.Ref.Close() }
func (a *Adapter) Resize(w, h int) { a.Ref.TryResizeUI(w, h) }
func (a *Adapter) Input(k string)  { a.Ref.Input(k) }

func (a *Adapter) handleRedraw(batches ...[]any) {
	for _, batch := range batches {
		if len(batch) == 0 {
			continue
		}

		eventName, ok := batch[0].(string)
		if !ok {
			continue
		}

		args := batch[1:]

		switch eventName {
		case "grid_line":
			a.handleGridLine(args)

		case "hl_attr_define":
			a.handleHlAttrDefine(args)

		case "grid_cursor_goto":
			// Expected: [grid_id, row, col]
			for _, arg := range args {
				posSlice, ok := arg.([]any)
				if ok && len(posSlice) >= 2 {
					// Use toInt() to handle uint64/int64
					r := len(posSlice) - 2
					c := len(posSlice) - 1
					a.CursorY = toInt(posSlice[r])
					a.CursorX = toInt(posSlice[c])
					a.Grid.SetCursor(a.CursorX, a.CursorY)
				}
			}

		case "flush":
			if a.OnFlush != nil {
				a.OnFlush()
			}
		}
	}
}

func (a *Adapter) handleGridLine(args []any) {
	for _, arg := range args {
		lineParams, ok := arg.([]any)

		if !ok {
			continue
		}

		// Use toInt() everywhere
		// row := int(lineParams[1].(int64)) -> CRASH
		row := toInt(lineParams[1])
		colStart := toInt(lineParams[2])
		cells := lineParams[3].([]any)

		col := colStart
		currentStyleID := 0

		for _, cellData := range cells {
			cellSlice, ok := cellData.([]any)
			if !ok {
				continue
			}

			text := cellSlice[0].(string)

			if len(cellSlice) >= 2 {
				currentStyleID = toInt(cellSlice[1])
			}

			repeat := 1
			if len(cellSlice) >= 3 {
				repeat = toInt(cellSlice[2])
			}

			style := a.styles[currentStyleID]
			if currentStyleID == 0 {
				style = style.Foreground(tcell.ColorWhite)
			}

			mainRune := ' '
			runes := []rune(text)
			if len(runes) > 0 {
				mainRune = runes[0]
			}

			for i := 0; i < repeat; i++ {
				a.Grid.SetContent(col, row, mainRune, style)
				col++
			}
		}
	}
}

func (a *Adapter) handleHlAttrDefine(args []any) {
	for _, arg := range args {
		params := arg.([]any)
		id := toInt(params[0])
		rgbAttrs := params[1].(map[string]any)

		style := tcell.StyleDefault

		if fg, ok := rgbAttrs["foreground"]; ok {
			c := toInt(fg)
			style = style.Foreground(tcell.NewRGBColor(
				int32((c>>16)&0xff),
				int32((c>>8)&0xff),
				int32(c&0xff),
			))
		}

		if bg, ok := rgbAttrs["background"]; ok {
			c := toInt(bg)
			style = style.Background(tcell.NewRGBColor(
				int32((c>>16)&0xff),
				int32((c>>8)&0xff),
				int32(c&0xff),
			))
		}

		if val, ok := rgbAttrs["bold"]; ok && val.(bool) {
			style = style.Bold(true)
		}
		if val, ok := rgbAttrs["reverse"]; ok && val.(bool) {
			style = style.Reverse(true)
		}

		a.styles[id] = style
	}
}
