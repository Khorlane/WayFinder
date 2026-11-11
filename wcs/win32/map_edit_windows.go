package win32

import (
	"fmt"
	"math"
	"slices"
	"syscall"

	"WayFinder/wmr"
)

const (
	wmCommand        = 0x0111
	wmMouseMove      = 0x0200
	wmLButtonDown    = 0x0201
	wmLButtonUp      = 0x0202
	wmCaptureChanged = 0x0215
	ssNotify         = 0x0100
	swShow           = 5
	mkShift          = 0x0004

	mapEditID   = 1001
	mapSaveID   = 1002
	mapCancelID = 1003
)

var (
	hMapEdit, hMapSave, hMapCancel uintptr
	commonControls                 = syscall.NewLazyDLL("comctl32.dll")
	procSetWindowSubclass          = commonControls.NewProc("SetWindowSubclass")
	procDefSubclassProc            = commonControls.NewProc("DefSubclassProc")
	procSetCapture                 = user32.NewProc("SetCapture")
	procReleaseCapture             = user32.NewProc("ReleaseCapture")
)

type mapEditor struct {
	active  *wmr.Correction
	failure bool
	draft   liveMapView
	status  string
	grid    mapGrid
	drag    *mapRoomDrag
}

type mapRoomDrag struct {
	index      int
	rooms      []wmr.Room
	pointer    point
	grid       mapGrid
	wholeLane  bool
	horizontal bool
	axisChosen bool
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
	hMapSave = create("Save correction", mapSaveID)
	hMapCancel = create("Skip correction", mapCancelID)
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
	procInvalidateRect.Call(hWMP, 0, 0)
}

func (s *liveShell) openMapEditor(correction *wmr.Correction, failure bool) {
	e := &s.mapEditor
	e.endDrag()
	e.active, e.failure = correction, failure
	e.draft = liveMapView{layout: correction.Draft, title: s.mapView.title}
	e.draft.layout.Rooms = slices.Clone(correction.Draft.Rooms)
	e.status = "Drag a room. Shift-drag moves its whole row or column."
	if failure {
		e.status += " The yellow room is the destination."
		setWindowText(hMapSave, "Save correction")
		setWindowText(hMapCancel, "Skip correction")
	} else {
		setWindowText(hMapSave, "Save map")
		setWindowText(hMapCancel, "Cancel")
	}
	e.grid = mapGrid{}
	procEnableWindow.Call(hWIC, 0)
	s.updateCorrectionControls()
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
	case mapSaveID:
		if e.active != nil {
			applied, err := e.active.Submit(e.draft.layout.Rooms)
			if err != nil {
				e.status = fmt.Sprintf("Edit needs changes. %v", err)
			} else {
				s.mapView.layout = applied
				if e.failure {
					s.mapView.err = nil
				}
				s.closeMapEditor()
			}
			procInvalidateRect.Call(hWMP, 0, 0)
		}
	case mapCancelID:
		if e.active != nil {
			e.active.Discard()
			s.closeMapEditor()
		}
	default:
		return false
	}
	return true
}

func (s *liveShell) closeMapEditor() {
	e := &s.mapEditor
	e.endDrag()
	e.active, e.draft, e.grid = nil, liveMapView{}, mapGrid{}
	s.updateCorrectionControls()
	if s.connected {
		procEnableWindow.Call(hWIC, 1)
		procSetFocus.Call(hWIC)
	}
}

func (e *mapEditor) endDrag() {
	if e.drag != nil {
		e.drag = nil
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
	}
	e.grid = e.draft.draw(item.dc, item.bounds, s.gmcpStatus, e.status, frozen)
}

func mapPanelProc(hwnd, message, wParam, lParam, _, _ uintptr) uintptr {
	if liveWindow != nil {
		e := &liveWindow.mapEditor
		if e.active != nil {
			p := point{int32(int16(lParam)), int32(int16(lParam >> 16))}
			switch message {
			case wmLButtonDown:
				if e.startDrag(p, wParam&mkShift != 0) {
					procSetCapture.Call(hwnd)
				}
				return 0
			case wmMouseMove:
				if e.drag != nil {
					e.moveDrag(p)
					return 0
				}
			case wmLButtonUp:
				if e.drag != nil {
					e.moveDrag(p)
					e.drag = nil
					procReleaseCapture.Call()
					procInvalidateRect.Call(hwnd, 0, 0)
					return 0
				}
			case wmCaptureChanged:
				if e.drag != nil {
					copy(e.draft.layout.Rooms, e.drag.rooms)
					e.drag = nil
					procInvalidateRect.Call(hwnd, 0, 0)
				}
			}
		}
	}
	result, _, _ := procDefSubclassProc.Call(hwnd, message, wParam, lParam)
	return result
}

func (e *mapEditor) startDrag(p point, wholeLane bool) bool {
	if e.grid.step == 0 {
		return false
	}
	index, distance := -1, math.MaxFloat64
	for i, room := range e.draft.layout.Rooms {
		at := e.grid.position(room)
		d := max(math.Abs(float64(p.X-at.X)), math.Abs(float64(p.Y-at.Y)))
		if d <= float64(e.grid.roomHalf()+2) && (d < distance || (d == distance && room.ID == e.draft.layout.Current)) {
			index, distance = i, d
		}
	}
	if index < 0 {
		return false
	}
	e.drag = &mapRoomDrag{
		index: index, rooms: slices.Clone(e.draft.layout.Rooms), pointer: p, grid: e.grid, wholeLane: wholeLane,
	}
	return true
}

func (e *mapEditor) moveDrag(p point) {
	d := e.drag
	area := d.grid.area
	p.X = max(area.Left, min(area.Right-1, p.X))
	p.Y = max(area.Top, min(area.Bottom-1, p.Y))
	columns := (float64(p.X) - float64(d.pointer.X)) / d.grid.step
	rows := (float64(p.Y) - float64(d.pointer.Y)) / d.grid.step
	if !d.axisChosen {
		if max(math.Abs(columns), math.Abs(rows)) < 0.5 {
			return
		}
		d.axisChosen, d.horizontal = true, math.Abs(columns) >= math.Abs(rows)
	}
	anchor := d.rooms[d.index]
	delta := int(math.Round(rows))
	if d.horizontal {
		delta = int(math.Round(columns))
	}
	for i, room := range d.rooms {
		sameLane := (d.horizontal && room.C == anchor.C) || (!d.horizontal && room.R == anchor.R)
		if i != d.index && !(d.wholeLane && sameLane) {
			continue
		}
		if d.horizontal {
			room.C += delta
		} else {
			room.R += delta
		}
		e.draft.layout.Rooms[i] = room
	}
	procInvalidateRect.Call(hWMP, 0, 0)
}
