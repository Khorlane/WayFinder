package win32

import (
	"syscall"
	"unsafe"
)

const (
	className  = "WayFinderWCSStubWindowClass"
	windowName = "WayFinder — WayFinder Client Shell (WCS)"
)

const (
	cwUseDefault int32 = -2147483648

	swShowDefault = 10

	wmCreate  = 0x0001
	wmDestroy = 0x0002
	wmSize    = 0x0005

	wsChild        = 0x40000000
	wsVisible      = 0x10000000
	wsBorder       = 0x00800000
	wsCaption      = 0x00C00000
	wsSysMenu      = 0x00080000
	wsMinimizeBox  = 0x00020000
	wsClipSiblings = 0x04000000
	wsClipChildren = 0x02000000

	wsOverlapped = 0x00000000
	wsTiled      = wsOverlapped

	wsOverlappedWindow = wsTiled | wsCaption | wsSysMenu | wsMinimizeBox

	ssLeft = 0x00000000

	colorWindow  = 5
	colorBtnFace = 15
	idcArrow     = 32512
	imageIcon    = 1
	imageCursor  = 2
	lrShared     = 0x00008000
	winW         = 1120
	winH         = 760
)

type point struct {
	X int32
	Y int32
}

type msg struct {
	HWnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      point
}

type rect struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type wndClassEx struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

var (
	hMainWnd uintptr
	hWOV     uintptr
	hWIC     uintptr
	hWMP     uintptr
	hWOVLbl  uintptr
	hWICLbl  uintptr
	hWMPLbl  uintptr
)

func RunWCS() {
	hInstance := getModuleHandle()
	classNamePtr := syscall.StringToUTF16Ptr(className)
	windowNamePtr := syscall.StringToUTF16Ptr(windowName)
	staticClass := syscall.StringToUTF16Ptr("STATIC")

	hCursor := loadSystemResource(imageCursor, idcArrow)

	wc := wndClassEx{
		CbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HInstance:     hInstance,
		HCursor:       hCursor,
		HbrBackground: uintptr(colorWindow + 1),
		LpszClassName: classNamePtr,
	}

	if r, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		panic("RegisterClassExW failed")
	}

	clientRect := rect{Left: 0, Top: 0, Right: winW, Bottom: winH}
	procAdjustWindowRectEx.Call(
		uintptr(unsafe.Pointer(&clientRect)),
		uintptr(wsOverlappedWindow),
		0,
		0,
	)

	fullW := int32(clientRect.Right - clientRect.Left)
	fullH := int32(clientRect.Bottom - clientRect.Top)

	hMainWnd = createWindowEx(
		0,
		classNamePtr,
		windowNamePtr,
		wsOverlappedWindow|wsClipSiblings|wsClipChildren,
		cwUseDefault,
		cwUseDefault,
		fullW,
		fullH,
		0,
		0,
		hInstance,
		0,
	)
	if hMainWnd == 0 {
		panic("CreateWindowExW failed")
	}

	_ = staticClass

	procShowWindow.Call(hMainWnd, swShowDefault)
	procUpdateWindow.Call(hMainWnd)

	var m msg
	for {
		ret, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(ret) == -1 {
			panic("GetMessageW failed")
		}
		if ret == 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}
}

func wndProc(hwnd uintptr, message uint32, wParam, lParam uintptr) uintptr {
	switch message {
	case wmCreate:
		createPanels(hwnd)
		layoutPanels(hwnd)
		return 0
	case wmSize:
		layoutPanels(hwnd)
		return 0
	case wmDestroy:
		procPostQuitMessage.Call(0)
		return 0
	}

	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(message), wParam, lParam)
	return r
}

func createPanels(parent uintptr) {
	hInstance := getModuleHandle()
	staticClass := syscall.StringToUTF16Ptr("STATIC")

	panelStyle := uint32(wsChild | wsVisible | wsBorder)
	labelStyle := uint32(wsChild | wsVisible | ssLeft)

	hWOV = createWindowEx(0, staticClass, syscall.StringToUTF16Ptr(""), panelStyle, 0, 0, 0, 0, parent, 0, hInstance, 0)
	hWIC = createWindowEx(0, staticClass, syscall.StringToUTF16Ptr(""), panelStyle, 0, 0, 0, 0, parent, 0, hInstance, 0)
	hWMP = createWindowEx(0, staticClass, syscall.StringToUTF16Ptr(""), panelStyle, 0, 0, 0, 0, parent, 0, hInstance, 0)

	hWOVLbl = createWindowEx(0, staticClass, syscall.StringToUTF16Ptr("WOV\r\nWayFinder Output View\r\nMUD output / transcript area"), labelStyle, 0, 0, 0, 0, hWOV, 0, hInstance, 0)
	hWICLbl = createWindowEx(0, staticClass, syscall.StringToUTF16Ptr("WIC\r\nWayFinder Input Console\r\n> command entry"), labelStyle, 0, 0, 0, 0, hWIC, 0, hInstance, 0)
	hWMPLbl = createWindowEx(0, staticClass, syscall.StringToUTF16Ptr("WMP\r\nWayFinder Map Panel\r\nmap display area"), labelStyle, 0, 0, 0, 0, hWMP, 0, hInstance, 0)

	_ = colorBtnFace
}

func layoutPanels(hwnd uintptr) {
	if hWOV == 0 || hWIC == 0 || hWMP == 0 {
		return
	}

	var r rect
	if ok, _, _ := procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&r))); ok == 0 {
		return
	}

	clientW := int32(r.Right - r.Left)
	clientH := int32(r.Bottom - r.Top)

	const margin int32 = 10
	const gap int32 = 10

	innerW := clientW - margin*2
	innerH := clientH - margin*2
	if innerW < 100 || innerH < 100 {
		return
	}

	rightW := innerW * 32 / 100
	leftW := innerW - rightW - gap
	wicH := innerH * 24 / 100
	wovH := innerH - wicH - gap

	leftX := margin
	rightX := margin + leftW + gap
	topY := margin
	wicY := margin + wovH + gap

	moveWindow(hWOV, leftX, topY, leftW, wovH)
	moveWindow(hWIC, leftX, wicY, leftW, wicH)
	moveWindow(hWMP, rightX, topY, rightW, innerH)

	const labelPad int32 = 12
	moveWindow(hWOVLbl, labelPad, labelPad, leftW-labelPad*2, wovH-labelPad*2)
	moveWindow(hWICLbl, labelPad, labelPad, leftW-labelPad*2, wicH-labelPad*2)
	moveWindow(hWMPLbl, labelPad, labelPad, rightW-labelPad*2, innerH-labelPad*2)
}

func moveWindow(hwnd uintptr, x, y, width, height int32) {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	procMoveWindow.Call(hwnd, uintptr(x), uintptr(y), uintptr(width), uintptr(height), 1)
}

func createWindowEx(exStyle uint32, className, windowName *uint16, style uint32, x, y, width, height int32, parent, menu, instance, param uintptr) uintptr {
	r, _, _ := procCreateWindowExW.Call(
		uintptr(exStyle),
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowName)),
		uintptr(style),
		uintptr(x),
		uintptr(y),
		uintptr(width),
		uintptr(height),
		parent,
		menu,
		instance,
		param,
	)
	return r
}

func getModuleHandle() uintptr {
	r, _, _ := procGetModuleHandleW.Call(0)
	if r == 0 {
		panic("GetModuleHandleW failed")
	}
	return r
}

func loadSystemResource(resourceType uint32, resourceID uint32) uintptr {
	r, _, _ := procLoadImageW.Call(
		0,
		uintptr(resourceID),
		uintptr(resourceType),
		0,
		0,
		lrShared,
	)
	if r == 0 {
		panic("LoadImageW failed")
	}
	return r
}

func init() {
	_ = procDestroyWindow
}
