//go:build windows

package start

import (
	"os"
)

func Kill(process *os.Process) error {
	return process.Kill()
}