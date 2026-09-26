//go:build unix

package start

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func RawRun(gameExe string) *exec.Cmd {
	cmd := exec.Command(gameExe)
	cmd.Dir = filepath.Dir(gameExe)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return cmd
}