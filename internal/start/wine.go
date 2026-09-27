//go:build unix

package start

import (
	"os/exec"
	"path/filepath"
)

func WineRun(gameExe string) *exec.Cmd {
	cmd := exec.Command("wine", gameExe)
	cmd.Dir = filepath.Dir(gameExe)

	return cmd
}