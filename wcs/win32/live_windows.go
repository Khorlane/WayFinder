package win32

import (
	"context"
	"syscall"
	"unsafe"

	"WayFinder/weg"
	"WayFinder/wmr"
)

const (
	wsThickFrame  = 0x00040000
	wsMaximizeBox = 0x00010000
	wsVScroll     = 0x00200000
	wsDisabled    = 0x08000000

	esMultiline   = 0x0004
	esAutoVScroll = 0x0040
	esAutoHScroll = 0x0080
	esReadOnly    = 0x0800

	wmSetText       = 0x000C
	wmGetText       = 0x000D
	wmGetTextLength = 0x000E
	wmSetFont       = 0x0030
	wmKeyDown       = 0x0100
	wmTimer         = 0x0113
	wmVScroll       = 0x0115
	emSetSel        = 0x00B1
	emReplaceSel    = 0x00C2
	emSetLimitText  = 0x00C5
	emSetReadOnly   = 0x00CF
	emSetPassword   = 0x00CC

	emExLimitText   = 0x0435
	emSetBackground = 0x0443
	emSetCharFormat = 0x0444
	emGetTextLength = 0x045F
	cfmColor        = 0x40000000
	scfSelection    = 0x0001
	unicodeCodePage = 1200

	vkReturn          = 0x0D
	vkUp              = 0x26
	vkDown            = 0x28
	sbBottom          = 7
	ansiFixedFont     = 11
	liveTimer         = 1
	maxTranscriptSize = 128 * 1024
	maxCommandHistory = 100
)

type livePlayer func(context.Context, string, <-chan string, func(weg.LiveText), func(wmr.Layout, string, error, wmr.LayoutEditing), func()) error

type liveUpdate struct {
	text         weg.LiveText
	mapView      *liveMapView
	editing      wmr.LayoutEditing
	status       string
	connected    bool
	disconnected bool
}

type liveShell struct {
	ctx             context.Context
	address         string
	play            livePlayer
	commands        chan string
	updates         chan liveUpdate
	connected       bool
	masked          bool
	history         []string
	historyIndex    int
	commandDraft    string
	retainedCommand string
	mapView         liveMapView
	gmcpStatus      string
	mapEditor       mapEditor
	mapViewport     mapViewport
	beginMapEdit    func() *wmr.Correction
}

var liveWindow *liveShell

type charFormat struct {
	size, mask, effects uint32
	height, offset      int32
	textColor           uint32
	charset, family     byte
	faceName            [32]uint16
}

type textLengthOptions struct {
	flags, codePage uint32
}

// COLORREF values are ordered as default, ANSI normal colors, then bright colors.
var transcriptColors = [...]uint32{
	0x00DDDDDD,
	0x00404040, 0x004848CF, 0x0064B458, 0x0048B4CC,
	0x00D18A54, 0x00BA60B8, 0x00B8B050, 0x00C0C0C0,
	0x00808080, 0x006B6BFF, 0x0074E674, 0x0073F1FF,
	0x00FFA87A, 0x00FF80E8, 0x00F5F582, 0x00FFFFFF,
}

// RunLiveWCS hosts the transcript and command controls for a live game session.
func RunLiveWCS(address string, play livePlayer) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	liveWindow = &liveShell{
		ctx:        ctx,
		address:    address,
		play:       play,
		commands:   make(chan string, 16),
		updates:    make(chan liveUpdate, 32),
		gmcpStatus: weg.GMCPWaitingStatus,
	}
	defer func() { liveWindow = nil }()
	RunWCS()
}

func createLivePanels(parent uintptr) {
	if err := msftedit.Load(); err != nil {
		panic(err)
	}
	instance := getModuleHandle()
	transcriptClass := syscall.StringToUTF16Ptr("RICHEDIT50W")
	editClass := syscall.StringToUTF16Ptr("EDIT")
	empty := syscall.StringToUTF16Ptr("")
	const baseStyle = wsChild | wsVisible | wsBorder
	hWOV = createWindowEx(0, transcriptClass, empty, baseStyle|wsVScroll|esMultiline|esAutoVScroll|esReadOnly, 0, 0, 0, 0, parent, 0, instance, 0)
	hWIC = createWindowEx(0, editClass, empty, baseStyle|wsDisabled|esAutoHScroll, 0, 0, 0, 0, parent, 0, instance, 0)
	hWICLbl = createWindowEx(0, syscall.StringToUTF16Ptr("STATIC"), syscall.StringToUTF16Ptr("Command (Enter to send)"), wsChild|wsVisible, 0, 0, 0, 0, parent, 0, instance, 0)
	hWMP = createWindowEx(0, syscall.StringToUTF16Ptr("STATIC"), syscall.StringToUTF16Ptr("WayFinder Map Panel"), baseStyle|ssOwnerDraw|ssNotify, 0, 0, 0, 0, parent, 0, instance, 0)
	if hWOV == 0 || hWIC == 0 || hWICLbl == 0 || hWMP == 0 {
		panic("Could not create live game controls")
	}
	createCorrectionControls(parent, instance)
	font, _, _ := procGetStockObject.Call(ansiFixedFont)
	procSendMessageW.Call(hWOV, wmSetFont, font, 1)
	procSendMessageW.Call(hWIC, wmSetFont, font, 1)
	procSendMessageW.Call(hWICLbl, wmSetFont, font, 1)
	procSendMessageW.Call(hWOV, emSetBackground, 0, 0x001C1C1C)
	procSendMessageW.Call(hWOV, emExLimitText, 0, maxTranscriptSize*2)
	procSendMessageW.Call(hWIC, emSetLimitText, 4096, 0)
}

func layoutLivePanels(hwnd uintptr) {
	var client rect
	if ok, _, _ := procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&client))); ok == 0 {
		return
	}
	const margin int32 = 10
	const inputHeight int32 = 26
	const labelHeight int32 = 18
	width := client.Right - client.Left - 2*margin
	inputTop := client.Bottom - client.Top - margin - inputHeight
	labelTop := inputTop - labelHeight
	mapWidth := width / 3
	outputWidth := width - mapWidth - margin
	moveWindow(hWOV, margin, margin, outputWidth, labelTop-2*margin)
	mapLeft, mapHeight := margin+outputWidth+margin, labelTop-2*margin
	if liveWindow.beginMapEdit != nil || liveWindow.mapEditor.active != nil {
		const footerHeight int32 = 28
		mapHeight -= footerHeight
		const buttonGap int32 = 6
		buttonWidth := (mapWidth - 2*buttonGap) / 3
		moveWindow(hMapEdit, mapLeft, margin+mapHeight+4, mapWidth, 24)
		moveWindow(hMapSave, mapLeft, margin+mapHeight+4, buttonWidth, 24)
		moveWindow(hMapCancel, mapLeft+buttonWidth+buttonGap, margin+mapHeight+4, buttonWidth, 24)
		moveWindow(hMapExit, mapLeft+2*(buttonWidth+buttonGap), margin+mapHeight+4, mapWidth-2*(buttonWidth+buttonGap), 24)
	}
	moveWindow(hWMP, mapLeft, margin, mapWidth, mapHeight)
	moveWindow(hWICLbl, margin, labelTop, width, labelHeight)
	moveWindow(hWIC, margin, inputTop, width, inputHeight)
}

func (s *liveShell) start(hwnd uintptr) {
	appendTranscript(weg.LiveSpan{Text: "Connecting to " + s.address + "...\r\n\r\n"})
	if id, _, _ := procSetTimer.Call(hwnd, liveTimer, 30, 0); id == 0 {
		appendTranscript(weg.LiveSpan{Text: "Could not start the game display.\r\n"})
		return
	}
	go func() {
		err := s.play(s.ctx, s.address, s.commands, func(text weg.LiveText) {
			s.postUpdate(liveUpdate{text: text})
		}, func(layout wmr.Layout, title string, err error, editing wmr.LayoutEditing) {
			s.postUpdate(liveUpdate{mapView: &liveMapView{layout: layout, title: title, err: err}, editing: editing})
		}, func() {
			s.postUpdate(liveUpdate{connected: true})
		})
		message := "Disconnected from the game."
		if err != nil {
			message = err.Error()
		}
		s.postUpdate(liveUpdate{status: message, disconnected: true})
	}()
}

func (s *liveShell) postUpdate(update liveUpdate) {
	select {
	case s.updates <- update:
	case <-s.ctx.Done():
	}
}

func (s *liveShell) drainUpdates() {
	for range cap(s.updates) {
		select {
		case update := <-s.updates:
			appendTranscript(update.text.Spans...)
			if update.text.GMCPStatus != "" {
				s.gmcpStatus = update.text.GMCPStatus
				procInvalidateRect.Call(hWMP, 0, 0)
			}
			if update.mapView != nil {
				s.mapView = *update.mapView
				hadEditButton := s.beginMapEdit != nil
				s.beginMapEdit = update.editing.Begin
				switch {
				case update.editing.Failure != nil:
					s.openMapEditor(update.editing.Failure, true)
				case s.mapEditor.active != nil:
					s.closeMapEditor()
				case hadEditButton != (s.beginMapEdit != nil):
					s.updateCorrectionControls()
				}
				procInvalidateRect.Call(hWMP, 0, 0)
			}
			if update.connected {
				s.connected = true
				if s.mapEditor.active == nil {
					procEnableWindow.Call(hWIC, 1)
					procSetFocus.Call(hWIC)
				}
			}
			if update.disconnected {
				s.connected = false
				s.beginMapEdit = nil
				if s.mapEditor.active != nil {
					s.closeMapEditor()
				} else {
					s.updateCorrectionControls()
				}
				appendTranscript(weg.LiveSpan{Text: "\r\n\r\n" + update.status + "\r\n"})
				procEnableWindow.Call(hWIC, 0)
				setWindowText(hWIC, "")
			}
			if len(update.text.Spans) > 0 && s.masked != update.text.MaskInput {
				s.masked = update.text.MaskInput
				var mask uintptr
				if s.masked {
					mask = '*'
					if s.retainedCommand != "" && commandText() == s.retainedCommand {
						setWindowText(hWIC, "")
					}
					s.retainedCommand = ""
				}
				procSendMessageW.Call(hWIC, emSetPassword, mask, 0)
			}
		default:
			return
		}
	}
}

func (s *liveShell) submitCommand() {
	if s.mapEditor.active != nil {
		return
	}
	command := commandText()
	select {
	case s.commands <- command:
		if s.masked {
			setWindowText(hWIC, "")
		} else {
			s.retainedCommand = command
			procSendMessageW.Call(hWIC, emSetSel, 0, ^uintptr(0))
		}
		if !s.masked && command != "" {
			if len(s.history) == maxCommandHistory {
				copy(s.history, s.history[1:])
				s.history = s.history[:maxCommandHistory-1]
			}
			s.history = append(s.history, command)
		}
		s.historyIndex = len(s.history)
		s.commandDraft = ""
		if !s.masked {
			appendTranscript(weg.LiveSpan{Text: command + "\r\n"})
		}
	default:
		appendTranscript(weg.LiveSpan{Text: "\r\nCommand queue is busy. Press Enter again to retry.\r\n"})
	}
}

func (s *liveShell) recallCommand(step int) {
	next := s.historyIndex + step
	if s.masked || next < 0 || next > len(s.history) {
		return
	}
	if s.historyIndex == len(s.history) {
		s.commandDraft = commandText()
	}
	s.historyIndex = next
	text := s.commandDraft
	if next < len(s.history) {
		text = s.history[next]
	}
	setWindowText(hWIC, text)
	length, _, _ := procSendMessageW.Call(hWIC, wmGetTextLength, 0, 0)
	procSendMessageW.Call(hWIC, emSetSel, length, length)
}

func commandText() string {
	length, _, _ := procSendMessageW.Call(hWIC, wmGetTextLength, 0, 0)
	buffer := make([]uint16, length+1)
	procSendMessageW.Call(hWIC, wmGetText, uintptr(len(buffer)), uintptr(unsafe.Pointer(&buffer[0])))
	return syscall.UTF16ToString(buffer)
}

func appendTranscript(spans ...weg.LiveSpan) {
	if len(spans) == 0 {
		return
	}
	// Rich Edit selection offsets count each paragraph break as one character.
	options := textLengthOptions{codePage: unicodeCodePage}
	length, _, _ := procSendMessageW.Call(hWOV, emGetTextLength, uintptr(unsafe.Pointer(&options)), 0)
	procSendMessageW.Call(hWOV, emSetReadOnly, 0, 0)
	if length > maxTranscriptSize {
		remove := length - maxTranscriptSize/2
		procSendMessageW.Call(hWOV, emSetSel, 0, remove)
		procSendMessageW.Call(hWOV, emReplaceSel, 0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(""))))
		length -= remove
	}
	procSendMessageW.Call(hWOV, emSetSel, length, length)
	format := charFormat{mask: cfmColor}
	format.size = uint32(unsafe.Sizeof(format))
	for _, span := range spans {
		format.textColor = transcriptColors[span.Color]
		procSendMessageW.Call(hWOV, emSetCharFormat, scfSelection, uintptr(unsafe.Pointer(&format)))
		procSendMessageW.Call(hWOV, emReplaceSel, 0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(span.Text))))
	}
	procSendMessageW.Call(hWOV, emSetReadOnly, 1, 0)
	procSendMessageW.Call(hWOV, wmVScroll, sbBottom, 0)
}

func setWindowText(hwnd uintptr, text string) {
	procSendMessageW.Call(hwnd, wmSetText, 0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))))
}
