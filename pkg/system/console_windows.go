//go:build windows

package system

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	kernel32                    = windows.NewLazySystemDLL("kernel32.dll")
	procSetConsoleWindowInfo    = kernel32.NewProc("SetConsoleWindowInfo")
	procSetConsoleScreenBufSize = kernel32.NewProc("SetConsoleScreenBufferSize")
)

type coord struct {
	X int16
	Y int16
}

type smallRect struct {
	Left   int16
	Top    int16
	Right  int16
	Bottom int16
}

func TrySetConsoleSize(cols, rows int16) {
	h, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE)
	if err != nil || h == 0 {
		return
	}

	buf := coord{X: cols, Y: rows + 200}
	_, _, _ = procSetConsoleScreenBufSize.Call(uintptr(h), uintptr(*(*int32)(unsafe.Pointer(&buf))))
	rect := smallRect{Left: 0, Top: 0, Right: cols - 1, Bottom: rows - 1}
	_, _, _ = procSetConsoleWindowInfo.Call(uintptr(h), uintptr(1), uintptr(unsafe.Pointer(&rect)))
	_, _, _ = procSetConsoleScreenBufSize.Call(uintptr(h), uintptr(*(*int32)(unsafe.Pointer(&buf))))
}
