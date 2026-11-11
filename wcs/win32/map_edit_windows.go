package win32

import (
	"fmt"
	"maps"
	"math"
	"slices"
	"syscall"
	"unsafe"

	"WayFinder/wmr"
)

const (
	wmCommand        = 0x0111
	wmMouseMove      = 0x0200
	wmLButtonDown    = 0x0201
	wmLButtonUp      = 0x0202
	wmLButtonDouble  = 0x0203
	wmCaptureChanged = 0x0215
	ssNotify         = 0x0100
	swShow           = 5
	mkShift          = 0x0004
	mkControl        = 0x0008

	mapEditID   = 1001
	mapSaveID   = 1002
	mapCancelID = 1003
	mapExitID   = 1004

	selectionDragThreshold = 4
	mapEditHelp            = "Drag to box-select. Ctrl-drag adds a box and Ctrl-click toggles a room. Drag selected rooms to move them. Shift-drag moves a row or column."
)

var (
	hMapEdit, hMapSave, hMapCancel, hMapExit uintptr
	commonControls                           = syscall.NewLazyDLL("comctl32.dll")
	procSetWindowSubclass                    = commonControls.NewProc("SetWindowSubclass")
	procDefSubclassProc                      = commonControls.NewProc("DefSubclassProc")
	procSetCapture                           = user32.NewProc("SetCapture")
	procReleaseCapture                       = user32.NewProc("ReleaseCapture")
	procFrameRect                            = user32.NewProc("FrameRect")
)

type mapEditor struct {
	active   *wmr.Correction
	failure  bool
	draft    liveMapView
	baseline wmr.Layout
	status   string
	grid     mapGrid
	selected map[wmr.RoomID]bool
	drag     *mapRoomDrag
	marquee  *mapMarquee
}

type mapRoomDrag struct {
	index      int
	rooms      []wmr.Room
	pointer    point
	grid       mapGrid
	wholeLane  bool
	group      bool
	horizontal bool
	axisChosen bool
}

type mapMarquee struct {
	start, end point
	grid       mapGrid
	additive   bool
	dragged    bool
}

func (m mapMarquee) bounds() rect {
	return rect{min(m.start.X, m.end.X), min(m.start.Y, m.end.Y), max(m.start.X, m.end.X) + 1, max(m.start.Y, m.end.Y) + 1}
}

func (m mapMarquee) selection(rooms []wmr.Room, existing map[wmr.RoomID]bool) map[wmr.RoomID]bool {
	selected := make(map[wmr.RoomID]bool)
	if m.additive {
		maps.Copy(selected, existing)
	}
	bounds := m.bounds()
	for _, room := range rooms {
		if bounds.contains(m.grid.position(room)) {
			selected[room.ID] = true
		}
	}
	return selected
}

func createCorrectionControls(parent, instance uintptr) {
	button := syscall.StringToUTF16Ptr("BUTTON")
	create := func(title string, id uintptr) uintptr {
		hwnd := createWindowEx(0, button, syscall.StringToUTF16Ptr(title), wsChild, 0, 0, 0, 0, parent, id, instance, 0)
		if hwnd == 0 {
			panic("Could not create map correction controls")
		}
		font, _, _ := procGetStockObject.Call(defaultGUIFont)
		procSendMessageW.Call(hwnd, wmSetFont, font, 1)
		return hwnd
	}
	hMapEdit = create("Edit map", mapEditID)
	hMapSave = create("Save", mapSaveID)
	hMapCancel = create("Cancel", mapCancelID)
	hMapExit = create("Exit", mapExitID)
	if ok, _, _ := procSetWindowSubclass.Call(hWMP, syscall.NewCallback(mapPanelProc), 1, 0); ok == 0 {
		panic("Could not enable map dragging")
	}
}

func (s *liveShell) updateCorrectionControls() {
	e := &s.mapEditor
	show := func(hwnd uintptr, visible bool) {
		var command uintptr
		if visible {
			command = swShow
		}
		procShowWindow.Call(hwnd, command)
	}
	layoutLivePanels(hMainWnd)
	show(hMapEdit, s.beginMapEdit != nil && e.active == nil)
	show(hMapSave, e.active != nil)
	show(hMapCancel, e.active != nil)
	show(hMapExit, e.active != nil)
	procInvalidateRect.Call(hWMP, 0, 0)
}

func (s *liveShell) openMapEditor(correction *wmr.Correction, failure bool) {
	e := &s.mapEditor
	e.endDrag()
	e.active, e.failure = correction, failure
	e.baseline = correction.Draft
	e.draft = liveMapView{title: s.mapView.title}
	e.resetDraft()
	procEnableWindow.Call(hWIC, 0)
	s.updateCorrectionControls()
}

func (e *mapEditor) resetDraft() {
	e.draft.layout = e.baseline
	e.draft.layout.Rooms = slices.Clone(e.baseline.Rooms)
	e.status = mapEditHelp
	if e.failure {
		e.status += " The yellow room is the destination."
	}
	e.grid = mapGrid{}
	e.selected = nil
}

func (s *liveShell) editMapCommand(id uintptr) bool {
	e := &s.mapEditor
	switch id {
	case mapEditID:
		if e.active == nil && s.beginMapEdit != nil {
			if edit := s.beginMapEdit(); edit != nil {
				s.openMapEditor(edit, false)
			}
		}
	case mapSaveID, mapExitID:
		if e.active != nil {
			e.endDrag()
			applied, err := e.active.Submit(e.draft.layout.Rooms)
			if len(applied.Rooms) != 0 {
				s.mapView.layout = applied
				if e.failure {
					s.mapView.err = nil
				}
				e.failure = false
				e.baseline = applied
				e.draft.layout = applied
				e.draft.layout.Rooms = slices.Clone(applied.Rooms)
			}
			switch {
			case err != nil:
				e.status = fmt.Sprintf("Save failed. %v", err)
			case id == mapExitID:
				s.closeMapEditor()
			default:
				e.status = "Map saved. " + mapEditHelp
			}
			procInvalidateRect.Call(hWMP, 0, 0)
		}
	case mapCancelID:
		if e.active != nil {
			e.endDrag()
			e.resetDraft()
			procInvalidateRect.Call(hWMP, 0, 0)
		}
	default:
		return false
	}
	return true
}

func (s *liveShell) closeMapEditor() {
	e := &s.mapEditor
	e.endDrag()
	if e.active != nil {
		e.active.Close()
	}
	*e = mapEditor{}
	s.updateCorrectionControls()
	if s.connected {
		procEnableWindow.Call(hWIC, 1)
		procSetFocus.Call(hWIC)
	}
}

func (e *mapEditor) endDrag() {
	if e.drag != nil || e.marquee != nil {
		e.drag, e.marquee = nil, nil
		procReleaseCapture.Call()
	}
}

func (s *liveShell) drawMap(item *drawItem) {
	e := &s.mapEditor
	if e.active == nil {
		s.mapView.draw(item.dc, item.bounds, s.gmcpStatus, "", nil)
		return
	}
	var frozen *mapGrid
	if e.drag != nil {
		frozen = &e.drag.grid
	} else if e.marquee != nil {
		frozen = &e.marquee.grid
	}
	e.grid = e.draft.draw(item.dc, item.bounds, s.gmcpStatus, e.status, frozen)
	e.drawSelection(item.dc)
}

func (e *mapEditor) drawSelection(dc uintptr) {
	if e.grid.step == 0 || (len(e.selected) == 0 && e.marquee == nil) {
		return
	}
	brush, _, _ := procCreateSolidBrush.Call(0x008DE6BE)
	if brush == 0 {
		return
	}
	defer procDeleteObject.Call(brush)

	selected := e.selected
	if e.marquee != nil && (!e.marquee.additive || e.marquee.dragged) {
		bounds := e.marquee.bounds()
		procFrameRect.Call(dc, uintptr(unsafe.Pointer(&bounds)), brush)
		selected = e.marquee.selection(e.draft.layout.Rooms, selected)
	}
	half := e.grid.roomHalf() + 1
	for _, room := range e.draft.layout.Rooms {
		if selected[room.ID] {
			at := e.grid.position(room)
			bounds := rect{at.X - half, at.Y - half, at.X + half + 1, at.Y + half + 1}
			procFrameRect.Call(dc, uintptr(unsafe.Pointer(&bounds)), brush)
		}
	}
}

func mapPanelProc(hwnd, message, wParam, lParam, _, _ uintptr) uintptr {
	if liveWindow != nil {
		e := &liveWindow.mapEditor
		if e.active != nil {
			p := point{int32(int16(lParam)), int32(int16(lParam >> 16))}
			switch message {
			case wmLButtonDown, wmLButtonDouble:
				if e.startDrag(p, wParam&mkShift != 0, wParam&mkControl != 0) {
					procSetCapture.Call(hwnd)
				}
				return 0
			case wmMouseMove:
				if e.updateDrag(p) {
					procInvalidateRect.Call(hwnd, 0, 0)
					return 0
				}
			case wmLButtonUp:
				if e.updateDrag(p) {
					if e.marquee != nil {
						e.finishSelection()
					}
					e.endDrag()
					procInvalidateRect.Call(hwnd, 0, 0)
					return 0
				}
			case wmCaptureChanged:
				if e.drag != nil {
					copy(e.draft.layout.Rooms, e.drag.rooms)
				}
				if e.drag != nil || e.marquee != nil {
					e.drag, e.marquee = nil, nil
					procInvalidateRect.Call(hwnd, 0, 0)
				}
			}
		}
	}
	result, _, _ := procDefSubclassProc.Call(hwnd, message, wParam, lParam)
	return result
}

func (e *mapEditor) roomAt(p point) int {
	if e.grid.step == 0 || !e.grid.area.contains(p) {
		return -1
	}
	index, distance := -1, math.MaxFloat64
	for i, room := range e.draft.layout.Rooms {
		at := e.grid.position(room)
		d := max(math.Abs(float64(p.X-at.X)), math.Abs(float64(p.Y-at.Y)))
		preferred := room.ID == e.draft.layout.Current
		if index >= 0 && e.selected[room.ID] != e.selected[e.draft.layout.Rooms[index].ID] {
			preferred = e.selected[room.ID]
		}
		if d <= float64(e.grid.roomHalf()+2) && (d < distance || (d == distance && preferred)) {
			index, distance = i, d
		}
	}
	return index
}

func (e *mapEditor) toggleSelection(p point) bool {
	index := e.roomAt(p)
	if index < 0 {
		return false
	}
	id := e.draft.layout.Rooms[index].ID
	if e.selected[id] {
		delete(e.selected, id)
	} else {
		if e.selected == nil {
			e.selected = make(map[wmr.RoomID]bool)
		}
		e.selected[id] = true
	}
	return true
}

func (e *mapEditor) startDrag(p point, wholeLane, additive bool) bool {
	if e.grid.step == 0 || !e.grid.area.contains(p) {
		return false
	}
	index := e.roomAt(p)
	if additive || index < 0 {
		e.marquee = &mapMarquee{start: p, end: p, grid: e.grid, additive: additive}
		return true
	}
	group := !wholeLane && e.selected[e.draft.layout.Rooms[index].ID]
	if !group {
		e.selected = nil
	}
	e.drag = &mapRoomDrag{
		index: index, rooms: slices.Clone(e.draft.layout.Rooms), pointer: p, grid: e.grid, wholeLane: wholeLane, group: group,
	}
	return true
}

func (e *mapEditor) updateDrag(p point) bool {
	if box := e.marquee; box != nil {
		box.end = box.grid.area.clamp(p)
		dx, dy := p.X-box.start.X, p.Y-box.start.Y
		box.dragged = box.dragged || max(dx, -dx, dy, -dy) >= selectionDragThreshold
		return true
	}
	if e.drag != nil {
		e.moveDrag(e.drag.grid.area.clamp(p))
		return true
	}
	return false
}

func (e *mapEditor) finishSelection() {
	if e.marquee.additive && !e.marquee.dragged {
		e.toggleSelection(e.marquee.start)
		return
	}
	e.selected = e.marquee.selection(e.draft.layout.Rooms, e.selected)
}

func (e *mapEditor) moveDrag(p point) {
	d := e.drag
	columns := (float64(p.X) - float64(d.pointer.X)) / d.grid.step
	rows := (float64(p.Y) - float64(d.pointer.Y)) / d.grid.step
	if !d.group && !d.axisChosen {
		if max(math.Abs(columns), math.Abs(rows)) < 0.5 {
			return
		}
		d.axisChosen, d.horizontal = true, math.Abs(columns) >= math.Abs(rows)
	}
	anchor := d.rooms[d.index]
	rowDelta, columnDelta := int(math.Round(rows)), int(math.Round(columns))
	if !d.group {
		if d.horizontal {
			rowDelta = 0
		} else {
			columnDelta = 0
		}
	}
	for i, room := range d.rooms {
		sameLane := (d.horizontal && room.C == anchor.C) || (!d.horizontal && room.R == anchor.R)
		if i != d.index && !(d.wholeLane && sameLane) && !(d.group && e.selected[room.ID]) {
			continue
		}
		room.R += rowDelta
		room.C += columnDelta
		e.draft.layout.Rooms[i] = room
	}
}
