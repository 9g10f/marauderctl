//go:build unix

package start

import (
	"os/exec"
	"path/filepath"
)

// Returns the command for running a specified game using Wine
func WineRun(gameExe string) *exec.Cmd {
	cmd := exec.Command("wine", gameExe)
	cmd.Dir = filepath.Dir(gameExe)

	return cmd
}