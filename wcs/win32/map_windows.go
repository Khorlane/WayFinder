package win32

import (
	"fmt"
	"slices"
	"syscall"
	"unsafe"

	"WayFinder/wmr"
)

const (
	ssOwnerDraw           = 0x000D
	wmDrawItem            = 0x002B
	defaultGUIFont        = 17
	transparentBackground = 1
	dtNoPrefix            = 0x0800
	dtSingleLine          = 0x0020
	dtEndEllipsis         = 0x8000
	dtWordBreak           = 0x0010
	mapBackground         = 0x00282320
	mapExitColor          = 0x00F3D7AF
)

var mapExitOffsets = map[string]point{
	"N": {0, -1}, "NE": {1, -1}, "E": {1, 0}, "SE": {1, 1},
	"S": {0, 1}, "SW": {-1, 1}, "W": {-1, 0}, "NW": {-1, -1},
}

var procIntersectClipRect = gdi32.NewProc("IntersectClipRect")

type drawItem struct {
	controlType, controlID, itemID, action, state uint32
	window, dc                                    uintptr
	bounds                                        rect
	data                                          uintptr
}

type liveMapView struct {
	layout wmr.Layout
	title  string
	err    error
}

type mapGrid struct {
	left, top float64
	row, col  int
	step      float64
	area      rect
}

func (g mapGrid) position(room wmr.Room) point {
	return point{
		X: int32(g.left + float64(room.C-g.col)*g.step),
		Y: int32(g.top + float64(room.R-g.row)*g.step),
	}
}

func (g mapGrid) center() (float64, float64) {
	return (float64(g.area.Left) + float64(g.area.Right)) / 2, (float64(g.area.Top) + float64(g.area.Bottom)) / 2
}

func (g mapGrid) roomHalf() int32 {
	return max(1, min(int32(g.step/4), max(7, int32(g.step/5))))
}

func (r rect) contains(p point) bool {
	return p.X >= r.Left && p.X < r.Right && p.Y >= r.Top && p.Y < r.Bottom
}

func (r rect) clamp(p point) point {
	return point{max(r.Left, min(r.Right-1, p.X)), max(r.Top, min(r.Bottom-1, p.Y))}
}

func (v liveMapView) draw(dc uintptr, bounds rect, gmcpStatus, editStatus string, viewport *mapViewport, dragGrid *mapGrid, overlay func(uintptr, mapGrid)) mapGrid {
	background, _, _ := procCreateSolidBrush.Call(mapBackground)
	fillMapRect(dc, bounds, background)
	procDeleteObject.Call(background)
	if bounds.Right-bounds.Left < 40 || bounds.Bottom-bounds.Top < 90 {
		return mapGrid{}
	}

	saved, _, _ := procSaveDC.Call(dc)
	if saved == 0 {
		return mapGrid{}
	}
	defer procRestoreDC.Call(dc, saved)
	font, _, _ := procGetStockObject.Call(defaultGUIFont)
	procSelectObject.Call(dc, font)
	procSetBkMode.Call(dc, transparentBackground)
	procSetTextColor.Call(dc, 0x00E8E8E8)

	area := rect{bounds.Left + 12, bounds.Top + 10, bounds.Right - 12, bounds.Bottom - 10}
	line := rect{area.Left, area.Top, area.Right, area.Top + 20}
	heading := "Map"
	if editStatus != "" {
		heading = "Editing map"
	}
	drawMapText(dc, line, heading, dtSingleLine|dtEndEllipsis)
	line.Top, line.Bottom = line.Bottom, line.Bottom+20
	drawMapText(dc, line, gmcpStatus, dtSingleLine|dtEndEllipsis)
	line.Top, line.Bottom = line.Bottom, line.Bottom+20
	if v.title != "" {
		drawMapText(dc, line, v.title, dtSingleLine|dtEndEllipsis)
		line.Top, line.Bottom = line.Bottom, line.Bottom+20
	}
	procSetTextColor.Call(dc, 0x00B8B8B8)
	if len(v.layout.Rooms) > 0 {
		drawMapText(dc, line, fmt.Sprintf("Rooms on map: %d", len(v.layout.Rooms)), dtSingleLine|dtEndEllipsis)
		line.Top, line.Bottom = line.Bottom, line.Bottom+20
	}
	if editStatus != "" {
		line.Bottom = min(area.Bottom, line.Top+64)
		drawMapText(dc, line, editStatus, dtWordBreak|dtEndEllipsis)
	} else if v.err != nil || len(v.layout.Rooms) == 0 {
		message := "Waiting for a room description."
		if v.err != nil {
			message = v.err.Error()
		}
		line.Bottom = min(area.Bottom, line.Top+54)
		drawMapText(dc, line, message, dtWordBreak|dtEndEllipsis)
	}
	area.Top = line.Bottom + 6
	if len(v.layout.Rooms) > 0 && area.Bottom-area.Top >= 12 {
		grid := viewport.project(fitMapGrid(v.layout.Rooms, area, editStatus != ""))
		if dragGrid != nil {
			grid = *dragGrid
		}
		procIntersectClipRect.Call(dc, uintptr(area.Left), uintptr(area.Top), uintptr(area.Right), uintptr(area.Bottom))
		v.drawRooms(dc, grid)
		if overlay != nil {
			overlay(dc, grid)
		}
		return grid
	}
	return mapGrid{}
}

func fitMapGrid(rooms []wmr.Room, area rect, editing bool) mapGrid {
	first := rooms[0]
	minRow, maxRow, minCol, maxCol := first.R, first.R, first.C, first.C
	for _, room := range rooms[1:] {
		minRow, maxRow = min(minRow, room.R), max(maxRow, room.R)
		minCol, maxCol = min(minCol, room.C), max(maxCol, room.C)
	}
	if editing {
		minRow, maxRow = minRow-1, maxRow+1
		minCol, maxCol = minCol-1, maxCol+1
	}
	width, height := float64(area.Right-area.Left), float64(area.Bottom-area.Top)
	step := min(36.0, width/float64(maxCol-minCol+1), height/float64(maxRow-minRow+1))
	left := float64(area.Left) + (width-float64(maxCol-minCol)*step)/2
	top := float64(area.Top) + (height-float64(maxRow-minRow)*step)/2
	return mapGrid{left: left, top: top, row: minRow, col: minCol, step: step, area: area}
}

func (v liveMapView) drawRooms(dc uintptr, grid mapGrid) {
	positions := make(map[wmr.RoomID]point, len(v.layout.Rooms))
	for _, room := range v.layout.Rooms {
		positions[room.ID] = grid.position(room)
	}

	pen, _, _ := procCreatePen.Call(0, 1, 0x008D8270)
	if pen != 0 {
		previous, _, _ := procSelectObject.Call(dc, pen)
		for _, link := range v.layout.Links {
			drawMapLine(dc, positions[link[0]], positions[link[1]])
		}
		procSelectObject.Call(dc, previous)
		procDeleteObject.Call(pen)
	}

	half := grid.roomHalf()
	// Leave a gap between ticks on neighboring boxes even at reduced scales.
	tickLength := min(max(6, int32(grid.step/6)), int32(grid.step)/5, (int32(grid.step)-2*half-2)/2)
	if tickLength > 0 {
		v.drawCompassExits(dc, positions, half, tickLength)
	}

	brush, _, _ := procCreateSolidBrush.Call(0x00D2AA7E)
	for _, room := range v.layout.Rooms {
		at := positions[room.ID]
		fillMapRect(dc, rect{at.X - half, at.Y - half, at.X + half + 1, at.Y + half + 1}, brush)
	}
	procDeleteObject.Call(brush)
	if at, ok := positions[v.layout.Current]; ok {
		active, _, _ := procCreateSolidBrush.Call(0x0074D8F7)
		fillMapRect(dc, rect{at.X - half, at.Y - half, at.X + half + 1, at.Y + half + 1}, active)
		procDeleteObject.Call(active)
	}
	if half >= 2 {
		v.drawVerticalExits(dc, positions, half)
	}
}

func (v liveMapView) drawCompassExits(dc uintptr, positions map[wmr.RoomID]point, half, length int32) {
	pen, _, _ := procCreatePen.Call(0, 1, mapExitColor)
	if pen == 0 {
		return
	}
	previous, _, _ := procSelectObject.Call(dc, pen)
	for _, room := range v.layout.Rooms {
		at := positions[room.ID]
		for _, dir := range v.layout.ExitDirections[room.ID] {
			offset, ok := mapExitOffsets[dir]
			if !ok {
				continue
			}
			from := point{at.X + offset.X*half, at.Y + offset.Y*half}
			to := point{from.X + offset.X*(length+1), from.Y + offset.Y*(length+1)}
			drawMapLine(dc, from, to)
		}
	}
	procSelectObject.Call(dc, previous)
	procDeleteObject.Call(pen)
}

func (v liveMapView) drawVerticalExits(dc uintptr, positions map[wmr.RoomID]point, half int32) {
	pen, _, _ := procCreatePen.Call(0, 1, mapBackground)
	if pen == 0 {
		return
	}
	previous, _, _ := procSelectObject.Call(dc, pen)
	radius, arm := min(half, max(4, half/2)), min(half-1, max(2, half/3))
	for _, room := range v.layout.Rooms {
		dirs := v.layout.ExitDirections[room.ID]
		up, down := slices.Contains(dirs, "U"), slices.Contains(dirs, "D")
		if !up && !down {
			continue
		}
		at := positions[room.ID]
		top, bottom := point{at.X, at.Y - radius}, point{at.X, at.Y + radius}
		drawMapLine(dc, top, point{bottom.X, bottom.Y + 1})
		if up {
			drawMapLine(dc, point{at.X - arm, top.Y + arm}, top)
			drawMapLine(dc, point{at.X + arm, top.Y + arm}, top)
		}
		if down {
			drawMapLine(dc, point{at.X - arm, bottom.Y - arm}, bottom)
			drawMapLine(dc, point{at.X + arm, bottom.Y - arm}, bottom)
		}
	}
	procSelectObject.Call(dc, previous)
	procDeleteObject.Call(pen)
}

func drawMapLine(dc uintptr, from, to point) {
	procMoveToEx.Call(dc, uintptr(from.X), uintptr(from.Y), 0)
	procLineTo.Call(dc, uintptr(to.X), uintptr(to.Y))
}

func fillMapRect(dc uintptr, bounds rect, brush uintptr) {
	procFillRect.Call(dc, uintptr(unsafe.Pointer(&bounds)), brush)
}

func drawMapText(dc uintptr, bounds rect, text string, flags uintptr) {
	procDrawTextW.Call(dc, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))), ^uintptr(0), uintptr(unsafe.Pointer(&bounds)), flags|dtNoPrefix)
}
