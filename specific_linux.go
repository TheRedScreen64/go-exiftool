//go:build linux

package exiftool

import (
	"os/exec"
)

func hideWindow(cmd *exec.Cmd) {
	// No-op for Linux
}
