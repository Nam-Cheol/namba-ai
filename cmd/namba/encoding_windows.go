//go:build windows

package main

import "syscall"

func configureUTF8Output() {
	const utf8CodePage = 65001

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setConsoleOutputCP := kernel32.NewProc("SetConsoleOutputCP")
	setConsoleCP := kernel32.NewProc("SetConsoleCP")

	_, _, _ = setConsoleOutputCP.Call(uintptr(utf8CodePage))
	_, _, _ = setConsoleCP.Call(uintptr(utf8CodePage))
}
