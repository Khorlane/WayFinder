package win32

import (
	"math"
	"unsafe"
)

const (
	wmRButtonDown = 0x0204
	wmRButtonUp   = 0x0205
	wmMouseWheel  = 0x020A
	wheelNotch    = 120
)

var procScreenToClient = user32.NewProc("ScreenToClient")

func mouseMessagePoint(lParam uintptr) point {
	return point{int32(int16(lParam)), int32(int16(lParam >> 16))}
}

// The viewport stores a view center in grid coordinates. A zero step fits
// the map to the panel until the player zooms or pans.
type mapViewport struct {
	row, col float64
	step     float64
	fitStep  float64
	grid     mapGrid
	pan      *mapPan
}

type mapPan struct {
	start point
	grid  mapGrid
}

func (v *mapViewport) project(fitted mapGrid) mapGrid {
	v.fitStep = fitted.step
	if v.step == 0 {
		return fitted
	}
	x, y := fitted.center()
	fitted.step = v.step
	fitted.left = x + (float64(fitted.col)-v.col)*v.step
	fitted.top = y + (float64(fitted.row)-v.row)*v.step
	return fitted
}

func (v *mapViewport) setGrid(grid mapGrid) {
	x, y := grid.center()
	v.col = float64(grid.col) + (x-grid.left)/grid.step
	v.row = float64(grid.row) + (y-grid.top)/grid.step
	v.step, v.grid = grid.step, grid
}

func (v *mapViewport) zoom(p point, delta int) {
	grid := v.grid
	minimum := min(2.0, v.fitStep/2)
	step := max(minimum, min(128, grid.step*math.Pow(1.2, float64(delta)/wheelNotch)))
	if step == grid.step {
		return
	}
	ratio := step / grid.step
	x, y := float64(p.X), float64(p.Y)
	grid.left = x + (grid.left-x)*ratio
	grid.top = y + (grid.top-y)*ratio
	grid.step = step
	v.setGrid(grid)
}

func (v *mapViewport) movePan(p point) {
	pan := v.pan
	if p == pan.start && v.grid == pan.grid {
		return
	}
	grid := pan.grid
	grid.left += float64(p.X - pan.start.X)
	grid.top += float64(p.Y - pan.start.Y)
	v.setGrid(grid)
}

func (s *liveShell) canNavigateMap() bool {
	e := &s.mapEditor
	return e.active == nil || (len(e.selected) == 0 && e.drag == nil && e.marquee == nil)
}

func (s *liveShell) zoomMapAtScreen(p point, delta int) bool {
	if hWMP == 0 {
		return false
	}
	ok, _, _ := procScreenToClient.Call(hWMP, uintptr(unsafe.Pointer(&p)))
	v := &s.mapViewport
	if ok == 0 || v.grid.step == 0 || !v.grid.area.contains(p) {
		return false
	}
	if s.canNavigateMap() && v.pan == nil && delta != 0 {
		v.zoom(p, delta)
		s.mapEditor.grid = v.grid
		procInvalidateRect.Call(hWMP, 0, 0)
	}
	return true
}

func (s *liveShell) endMapPan() {
	if s.mapViewport.pan != nil {
		s.mapViewport.pan = nil
		procReleaseCapture.Call()
	}
}
