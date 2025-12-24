package engine

func (e *Engine) Resize(w, h int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.CodeGrid.Resize(w, h)
	e.RenderGrid.Resize(w, h)
}
