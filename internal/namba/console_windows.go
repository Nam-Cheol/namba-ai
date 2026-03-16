//go:build windows

package namba

import (
	"io"
	"os"
	"syscall"
	"unsafe"
)

const (
	consoleEchoInput             = 0x0004
	consoleLineInput             = 0x0002
	consoleProcessedInput        = 0x0001
	consoleVirtualTerminalInput  = 0x0200
	consoleVirtualTerminalOutput = 0x0004
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode = kernel32.NewProc("SetConsoleMode")
)

func enableRawConsoleInput(file *os.File) (func(), error) {
	handle := syscall.Handle(file.Fd())
	mode, err := getConsoleMode(handle)
	if err != nil {
		return nil, err
	}

	rawMode := mode &^ (consoleEchoInput | consoleLineInput | consoleProcessedInput)
	rawMode |= consoleVirtualTerminalInput
	if err := setConsoleMode(handle, rawMode); err != nil {
		return nil, err
	}

	return func() {
		_ = setConsoleMode(handle, mode)
	}, nil
}

func enableVirtualTerminalOutput(out io.Writer) func() {
	file, ok := out.(*os.File)
	if !ok {
		return func() {}
	}

	handle := syscall.Handle(file.Fd())
	mode, err := getConsoleMode(handle)
	if err != nil {
		return func() {}
	}

	updated := mode | consoleVirtualTerminalOutput
	if err := setConsoleMode(handle, updated); err != nil {
		return func() {}
	}

	return func() {
		_ = setConsoleMode(handle, mode)
	}
}

func getConsoleMode(handle syscall.Handle) (uint32, error) {
	var mode uint32
	result, _, err := procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
	if result == 0 {
		return 0, err
	}
	return mode, nil
}

func setConsoleMode(handle syscall.Handle, mode uint32) error {
	result, _, err := procSetConsoleMode.Call(uintptr(handle), uintptr(mode))
	if result == 0 {
		return err
	}
	return nil
}