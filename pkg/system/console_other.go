//go:build !windows

package system

func TrySetConsoleSize(cols, rows int16) {}
