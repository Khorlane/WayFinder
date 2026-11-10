package win32

import (
	"context"
	"syscall"
	"unsafe"

	"WayFinder/weg"
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
	emSetSel        = 0x00B1
	emScrollCaret   = 0x00B7
	emReplaceSel    = 0x00C2
	emSetLimitText  = 0x00C5
	emSetPassword   = 0x00CC

	vkReturn          = 0x0D
	ansiFixedFont     = 11
	liveTimer         = 1
	maxTranscriptSize = 128 * 1024
)

type livePlayer func(context.Context, string, <-chan string, func(weg.LiveText), func()) error

type liveUpdate struct {
	text         weg.LiveText
	connected    bool
	disconnected bool
}

type liveShell struct {
	ctx      context.Context
	address  string
	play     livePlayer
	commands chan string
	updates  chan liveUpdate
	masked   bool
}

var liveWindow *liveShell

// RunLiveWCS hosts the transcript and command controls for a live game session.
func RunLiveWCS(address string, play livePlayer) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	liveWindow = &liveShell{
		ctx:      ctx,
		address:  address,
		play:     play,
		commands: make(chan string, 16),
		updates:  make(chan liveUpdate, 32),
	}
	defer func() { liveWindow = nil }()
	RunWCS()
}

func createLivePanels(parent uintptr) {
	instance := getModuleHandle()
	editClass := syscall.StringToUTF16Ptr("EDIT")
	empty := syscall.StringToUTF16Ptr("")
	const baseStyle = wsChild | wsVisible | wsBorder
	hWOV = createWindowEx(0, editClass, empty, baseStyle|wsVScroll|esMultiline|esAutoVScroll|esReadOnly, 0, 0, 0, 0, parent, 0, instance, 0)
	hWIC = createWindowEx(0, editClass, empty, baseStyle|wsDisabled|esAutoHScroll, 0, 0, 0, 0, parent, 0, instance, 0)
	hWICLbl = createWindowEx(0, syscall.StringToUTF16Ptr("STATIC"), syscall.StringToUTF16Ptr("Command (Enter to send)"), wsChild|wsVisible, 0, 0, 0, 0, parent, 0, instance, 0)
	if hWOV == 0 || hWIC == 0 || hWICLbl == 0 {
		panic("Could not create live game controls")
	}
	font, _, _ := procGetStockObject.Call(ansiFixedFont)
	procSendMessageW.Call(hWOV, wmSetFont, font, 1)
	procSendMessageW.Call(hWIC, wmSetFont, font, 1)
	procSendMessageW.Call(hWICLbl, wmSetFont, font, 1)
	procSendMessageW.Call(hWOV, emSetLimitText, maxTranscriptSize*2, 0)
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
	moveWindow(hWOV, margin, margin, width, labelTop-2*margin)
	moveWindow(hWICLbl, margin, labelTop, width, labelHeight)
	moveWindow(hWIC, margin, inputTop, width, inputHeight)
}

func (s *liveShell) start(hwnd uintptr) {
	appendTranscript("Connecting to " + s.address + "...\r\n\r\n")
	if id, _, _ := procSetTimer.Call(hwnd, liveTimer, 30, 0); id == 0 {
		appendTranscript("Could not start the game display.\r\n")
		return
	}
	go func() {
		err := s.play(s.ctx, s.address, s.commands, func(text weg.LiveText) {
			s.postUpdate(liveUpdate{text: text})
		}, func() {
			s.postUpdate(liveUpdate{connected: true})
		})
		message := "Disconnected from the game."
		if err != nil {
			message = err.Error()
		}
		s.postUpdate(liveUpdate{text: weg.LiveText{Text: "\r\n\r\n" + message + "\r\n"}, disconnected: true})
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
			appendTranscript(update.text.Text)
			if update.connected {
				procEnableWindow.Call(hWIC, 1)
				procSetFocus.Call(hWIC)
			}
			if update.disconnected {
				procEnableWindow.Call(hWIC, 0)
				setWindowText(hWIC, "")
			}
			if s.masked != update.text.MaskInput {
				s.masked = update.text.MaskInput
				var mask uintptr
				if s.masked {
					mask = '*'
				}
				procSendMessageW.Call(hWIC, emSetPassword, mask, 0)
			}
		default:
			return
		}
	}
}

func (s *liveShell) submitCommand() {
	length, _, _ := procSendMessageW.Call(hWIC, wmGetTextLength, 0, 0)
	buffer := make([]uint16, length+1)
	procSendMessageW.Call(hWIC, wmGetText, uintptr(len(buffer)), uintptr(unsafe.Pointer(&buffer[0])))
	select {
	case s.commands <- syscall.UTF16ToString(buffer):
		setWindowText(hWIC, "")
	default:
		appendTranscript("\r\nCommand queue is busy. Press Enter again to retry.\r\n")
	}
}

func appendTranscript(text string) {
	if text == "" {
		return
	}
	length, _, _ := procSendMessageW.Call(hWOV, wmGetTextLength, 0, 0)
	if length > maxTranscriptSize {
		remove := length - maxTranscriptSize/2
		procSendMessageW.Call(hWOV, emSetSel, 0, remove)
		procSendMessageW.Call(hWOV, emReplaceSel, 0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(""))))
		length -= remove
	}
	procSendMessageW.Call(hWOV, emSetSel, length, length)
	procSendMessageW.Call(hWOV, emReplaceSel, 0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))))
	procSendMessageW.Call(hWOV, emScrollCaret, 0, 0)
}

func setWindowText(hwnd uintptr, text string) {
	procSendMessageW.Call(hwnd, wmSetText, 0, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))))
}
