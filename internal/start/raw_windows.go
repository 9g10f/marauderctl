//go:build windows

package start

import (
	"os"
	"os/exec"
	"path/filepath"
)

func RawRun(gameExe string) *exec.Cmd {
	cmd := exec.Command(gameExe)
	cmd.Dir = filepath.Dir(gameExe)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}