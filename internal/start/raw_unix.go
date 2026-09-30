//go:build unix

package start

import (
	"os/exec"
	"path/filepath"
	"syscall"
)

// Returns the command for running a specified game natively
func RawRun(gameExe string) *exec.Cmd {
	cmd := exec.Command(gameExe)
	cmd.Dir = filepath.Dir(gameExe)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return cmd
}