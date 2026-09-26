//go:build unix

package start

import (
	"os"
	"os/exec"
	"path/filepath"
)

func WineRun(gameExe string) *exec.Cmd {
	cmd := exec.Command("wine", gameExe)
	cmd.Dir = filepath.Dir(gameExe)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}