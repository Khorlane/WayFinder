package win32

import (
	"fmt"
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
)

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

func (v liveMapView) draw(dc uintptr, bounds rect, gmcpStatus string) {
	background, _, _ := procCreateSolidBrush.Call(0x00282320)
	fillMapRect(dc, bounds, background)
	procDeleteObject.Call(background)
	if bounds.Right-bounds.Left < 40 || bounds.Bottom-bounds.Top < 90 {
		return
	}

	saved, _, _ := procSaveDC.Call(dc)
	if saved == 0 {
		return
	}
	defer procRestoreDC.Call(dc, saved)
	font, _, _ := procGetStockObject.Call(defaultGUIFont)
	procSelectObject.Call(dc, font)
	procSetBkMode.Call(dc, transparentBackground)
	procSetTextColor.Call(dc, 0x00E8E8E8)

	area := rect{bounds.Left + 12, bounds.Top + 10, bounds.Right - 12, bounds.Bottom - 10}
	line := rect{area.Left, area.Top, area.Right, area.Top + 20}
	drawMapText(dc, line, "Map", dtSingleLine|dtEndEllipsis)
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
	if v.err != nil || len(v.layout.Rooms) == 0 {
		message := "Waiting for a room description."
		if v.err != nil {
			message = v.err.Error()
		}
		line.Bottom = min(area.Bottom, line.Top+54)
		drawMapText(dc, line, message, dtWordBreak|dtEndEllipsis)
	}
	area.Top = line.Bottom + 6
	if len(v.layout.Rooms) > 0 && area.Bottom-area.Top >= 12 {
		v.drawRooms(dc, area)
	}
}

func (v liveMapView) drawRooms(dc uintptr, area rect) {
	first := v.layout.Rooms[0]
	minRow, maxRow, minCol, maxCol := first.R, first.R, first.C, first.C
	for _, room := range v.layout.Rooms[1:] {
		minRow, maxRow = min(minRow, room.R), max(maxRow, room.R)
		minCol, maxCol = min(minCol, room.C), max(maxCol, room.C)
	}
	width, height := float64(area.Right-area.Left), float64(area.Bottom-area.Top)
	step := min(36.0, width/float64(maxCol-minCol+1), height/float64(maxRow-minRow+1))
	left := float64(area.Left) + (width-float64(maxCol-minCol)*step)/2
	top := float64(area.Top) + (height-float64(maxRow-minRow)*step)/2
	positions := make(map[wmr.RoomID]point, len(v.layout.Rooms))
	for _, room := range v.layout.Rooms {
		positions[room.ID] = point{
			X: int32(left + float64(room.C-minCol)*step),
			Y: int32(top + float64(room.R-minRow)*step),
		}
	}

	pen, _, _ := procCreatePen.Call(0, 1, 0x008D8270)
	if pen != 0 {
		previous, _, _ := procSelectObject.Call(dc, pen)
		for _, link := range v.layout.Links {
			from, to := positions[link[0]], positions[link[1]]
			procMoveToEx.Call(dc, uintptr(from.X), uintptr(from.Y), 0)
			procLineTo.Call(dc, uintptr(to.X), uintptr(to.Y))
		}
		procSelectObject.Call(dc, previous)
		procDeleteObject.Call(pen)
	}

	brush, _, _ := procCreateSolidBrush.Call(0x00D2AA7E)
	half := max(1, min(7, int32(step/4)))
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
}

func fillMapRect(dc uintptr, bounds rect, brush uintptr) {
	procFillRect.Call(dc, uintptr(unsafe.Pointer(&bounds)), brush)
}

func drawMapText(dc uintptr, bounds rect, text string, flags uintptr) {
	procDrawTextW.Call(dc, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))), ^uintptr(0), uintptr(unsafe.Pointer(&bounds)), flags|dtNoPrefix)
}
