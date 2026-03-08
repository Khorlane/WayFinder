package win32

import "syscall"

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procAdjustWindowRectEx = user32.NewProc("AdjustWindowRectEx")
	procCreateWindowExW    = user32.NewProc("CreateWindowExW")
	procDefWindowProcW     = user32.NewProc("DefWindowProcW")
	procDestroyWindow      = user32.NewProc("DestroyWindow")
	procDispatchMessageW   = user32.NewProc("DispatchMessageW")
	procGetClientRect      = user32.NewProc("GetClientRect")
	procGetMessageW        = user32.NewProc("GetMessageW")
	procLoadImageW         = user32.NewProc("LoadImageW")
	procMoveWindow         = user32.NewProc("MoveWindow")
	procPostQuitMessage    = user32.NewProc("PostQuitMessage")
	procRegisterClassExW   = user32.NewProc("RegisterClassExW")
	procShowWindow         = user32.NewProc("ShowWindow")
	procTranslateMessage   = user32.NewProc("TranslateMessage")
	procUpdateWindow       = user32.NewProc("UpdateWindow")

	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)
