package main

import (
	"fmt"
	"syscall"
	"unsafe"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	procSetWindowsHookEx    = user32.NewProc("SetWindowsHookExW")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procGetMessage          = user32.NewProc("GetMessageW")
	procGetAsyncKeyState    = user32.NewProc("GetAsyncKeyState")
)

const (
	WH_KEYBOARD_LL = 13
	WM_KEYDOWN     = 0x0100
	VK_O           = 0x4F
	VK_CONTROL     = 0x11
)

// KBDLLHOOKSTRUCT contains information about a low-level keyboard input event
type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

type MSG struct {
	HWND    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct{ X, Y int32 }
}

var appInstance *App
var keyboardHook uintptr

// isKeyPressed checks if a key is currently pressed
func isKeyPressed(vk uintptr) bool {
	ret, _, _ := procGetAsyncKeyState.Call(vk)
	return ret&0x8000 != 0
}

// keyboardProc is the low-level keyboard hook callback
func keyboardProc(nCode int, wParam uintptr, lParam uintptr) uintptr {
	if nCode >= 0 && wParam == WM_KEYDOWN {
		kbStruct := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))
		// Check for Ctrl+O
		if kbStruct.VkCode == VK_O && isKeyPressed(VK_CONTROL) {
			if appInstance != nil {
				appInstance.ToggleWindow()
			}
		}
	}
	ret, _, _ := procCallNextHookEx.Call(keyboardHook, uintptr(nCode), wParam, lParam)
	return ret
}

// RegisterToggleHotkey registers Ctrl+O as a global hotkey using low-level keyboard hook
func (a *App) RegisterToggleHotkey() {
	appInstance = a

	go func() {
		// Create callback
		callback := syscall.NewCallback(keyboardProc)

		// Install the low-level keyboard hook
		ret, _, err := procSetWindowsHookEx.Call(
			WH_KEYBOARD_LL,
			callback,
			0,
			0,
		)
		if ret == 0 {
			fmt.Printf("Failed to install keyboard hook: %v\n", err)
			return
		}
		keyboardHook = ret
		fmt.Println("Installed low-level keyboard hook for Ctrl+O")

		// Message loop to keep the hook alive
		var msg MSG
		for {
			ret, _, _ := procGetMessage.Call(
				uintptr(unsafe.Pointer(&msg)),
				0, 0, 0,
			)
			if ret == 0 {
				break
			}
		}
	}()
}

// ToggleWindow toggles the window visibility
func (a *App) ToggleWindow() {
	a.windowVisible = !a.windowVisible
	if a.windowVisible {
		fmt.Println("Showing overlay (Ctrl+O)")
		a.showWindow()
	} else {
		fmt.Println("Hiding overlay (Ctrl+O)")
		a.hideWindow()
	}
}

func (a *App) showWindow() {
	if a.ctx != nil {
		// Use wails runtime to show
		go func() {
			// Small delay to ensure we're not in a Windows message loop
			if a.ctx != nil {
				wailsRuntime.WindowShow(a.ctx)
			}
		}()
	}
}

func (a *App) hideWindow() {
	if a.ctx != nil {
		go func() {
			if a.ctx != nil {
				wailsRuntime.WindowHide(a.ctx)
			}
		}()
	}
}
