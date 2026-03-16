//go:build !windows

package namba

import (
	"io"
	"os"
	"os/exec"
	"strings"
)

func enableRawConsoleInput(file *os.File) (func(), error) {
	stateCmd := exec.Command("stty", "-g")
	stateCmd.Stdin = file
	state, err := stateCmd.Output()
	if err != nil {
		return nil, err
	}

	rawCmd := exec.Command("stty", "raw", "-echo")
	rawCmd.Stdin = file
	if err := rawCmd.Run(); err != nil {
		return nil, err
	}

	original := strings.TrimSpace(string(state))
	return func() {
		restoreCmd := exec.Command("stty", original)
		restoreCmd.Stdin = file
		_ = restoreCmd.Run()
	}, nil
}

func enableVirtualTerminalOutput(_ io.Writer) func() {
	return func() {}
}
