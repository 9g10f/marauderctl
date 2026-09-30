//go:build unix

package disk

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// Returns the amount of space left on a device from it's root path
func GetAvailableSpace(path string) (int, error) {
	var stat unix.Statfs_t

	err := unix.Statfs(path, &stat)
	if err != nil {
		return 0, fmt.Errorf("GetAvailableSpace: Unable to inspect device: '%v'", path)
	}

	return int(stat.Bavail) * int(stat.Bsize), nil
}