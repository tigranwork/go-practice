//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procRegisterClassExW = user32.NewProc("RegisterClassExW")
	procCreateWindowExW  = user32.NewProc("CreateWindowExW")
	procDefWindowProcW   = user32.NewProc("DefWindowProcW")
	procDispatchMessageW = user32.NewProc("DispatchMessageW")
	procGetMessageW      = user32.NewProc("GetMessageW")
	procPostQuitMessage  = user32.NewProc("PostQuitMessage")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procLoadIconW        = user32.NewProc("LoadIconW")
	procLoadCursorW      = user32.NewProc("LoadCursorW")
	procMessageBoxW      = user32.NewProc("MessageBoxW")
	procShowWindow       = user32.NewProc("ShowWindow")
	procUpdateWindow     = user32.NewProc("UpdateWindow")
	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	CW_USEDEFAULT       = 0x80000000
	SW_SHOWDEFAULT      = 10
	WM_DESTROY          = 0x0002
	WM_COMMAND          = 0x0111
	BN_CLICKED          = 0
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	BS_PUSHBUTTON       = 0x00000000
	COLOR_WINDOW        = 5
	IDC_ARROW           = 32512
	IDI_APPLICATION     = 32512
	ID_BUTTON           = 1
)

type wndClassEx struct {
	cbSize        uint32
	style         uint32
	lpfnWndProc   uintptr
	cbClsExtra    int32
	cbWndExtra    int32
	hInstance     syscall.Handle
	hIcon         syscall.Handle
	hCursor       syscall.Handle
	hbrBackground syscall.Handle
	lpszMenuName  *uint16
	lpszClassName *uint16
	hIconSm       syscall.Handle
}

type point struct {
	x, y int32
}

type msg struct {
	hwnd    syscall.Handle
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      point
}

func main() {
	instance := getModuleHandle()
	className := syscall.StringToUTF16Ptr("GoButtonWindow")

	wnd := wndClassEx{
		cbSize:        uint32(unsafe.Sizeof(wndClassEx{})),
		lpfnWndProc:   syscall.NewCallback(windowProc),
		hInstance:     instance,
		hIcon:         loadIcon(instance, IDI_APPLICATION),
		hCursor:       loadCursor(0, IDC_ARROW),
		hbrBackground: syscall.Handle(COLOR_WINDOW + 1),
		lpszClassName: className,
	}

	registerClass(&wnd)

	windowTitle := syscall.StringToUTF16Ptr("Go Windows Button App")
	hwnd := createWindow(
		0,
		className,
		windowTitle,
		WS_OVERLAPPEDWINDOW|WS_VISIBLE,
		CW_USEDEFAULT,
		CW_USEDEFAULT,
		420,
		240,
		0,
		0,
		instance,
		0,
	)

	createButton(hwnd, instance)

	showWindow(hwnd, SW_SHOWDEFAULT)
	updateWindow(hwnd)

	var message msg
	for {
		ret := getMessage(&message, 0, 0, 0)
		if ret == 0 {
			break
		}
		translateMessage(&message)
		dispatchMessage(&message)
	}
}

func createButton(parent syscall.Handle, instance syscall.Handle) {
	className := syscall.StringToUTF16Ptr("BUTTON")
	buttonText := syscall.StringToUTF16Ptr("Click me")

	procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(buttonText)),
		WS_CHILD|WS_VISIBLE|BS_PUSHBUTTON,
		110,
		80,
		200,
		50,
		uintptr(parent),
		uintptr(ID_BUTTON),
		uintptr(instance),
		0,
	)
}

func windowProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_COMMAND:
		if lowWord(uint32(wParam)) == ID_BUTTON && highWord(uint32(wParam)) == BN_CLICKED {
			showMessage(hwnd, "you clicked button")
			return 0
		}
	case WM_DESTROY:
		postQuitMessage(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(
		uintptr(hwnd),
		uintptr(msg),
		wParam,
		lParam,
	)
	return ret
}

func showMessage(hwnd syscall.Handle, text string) {
	title := syscall.StringToUTF16Ptr("Button Clicked")
	body := syscall.StringToUTF16Ptr(text)
	procMessageBoxW.Call(
		uintptr(hwnd),
		uintptr(unsafe.Pointer(body)),
		uintptr(unsafe.Pointer(title)),
		0,
	)
}

func registerClass(wnd *wndClassEx) {
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(wnd)))
}

func createWindow(exStyle uint32, className, windowName *uint16, style uint32, x, y, width, height int32, parent, menu uintptr, instance syscall.Handle, param uintptr) syscall.Handle {
	hwnd, _, _ := procCreateWindowExW.Call(
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
		uintptr(instance),
		param,
	)
	return syscall.Handle(hwnd)
}

func getMessage(message *msg, hwnd syscall.Handle, msgFilterMin, msgFilterMax uint32) uint32 {
	ret, _, _ := procGetMessageW.Call(
		uintptr(unsafe.Pointer(message)),
		uintptr(hwnd),
		uintptr(msgFilterMin),
		uintptr(msgFilterMax),
	)
	return uint32(ret)
}

func translateMessage(message *msg) {
	procTranslateMessage.Call(uintptr(unsafe.Pointer(message)))
}

func dispatchMessage(message *msg) {
	procDispatchMessageW.Call(uintptr(unsafe.Pointer(message)))
}

func postQuitMessage(exitCode int32) {
	procPostQuitMessage.Call(uintptr(exitCode))
}

func showWindow(hwnd syscall.Handle, cmdShow int32) {
	procShowWindow.Call(uintptr(hwnd), uintptr(cmdShow))
}

func updateWindow(hwnd syscall.Handle) {
	procUpdateWindow.Call(uintptr(hwnd))
}

func loadCursor(instance syscall.Handle, cursorID int32) syscall.Handle {
	cursor, _, _ := procLoadCursorW.Call(
		uintptr(instance),
		uintptr(cursorID),
	)
	return syscall.Handle(cursor)
}

func loadIcon(instance syscall.Handle, iconID int32) syscall.Handle {
	icon, _, _ := procLoadIconW.Call(
		uintptr(instance),
		uintptr(iconID),
	)
	return syscall.Handle(icon)
}

func getModuleHandle() syscall.Handle {
	handle, _, _ := procGetModuleHandleW.Call(0)
	return syscall.Handle(handle)
}

func lowWord(value uint32) uint16 {
	return uint16(value & 0xFFFF)
}

func highWord(value uint32) uint16 {
	return uint16(value >> 16)
}
