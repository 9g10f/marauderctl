//go:build windows

package disk

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// Returns the amount of space left on a device from it's root path
func GetAvailableSpace(path string) (int, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, fmt.Errorf("GetAvailableSpace: Unable to convert path to pointer: '%v'", path)
	}

	var freeBytes uint64
	var totalBytes uint64
	var totalFreeBytes uint64

	err = windows.GetDiskFreeSpaceEx(p, &freeBytes, &totalBytes, &totalFreeBytes)
	if err != nil {
		return 0, fmt.Errorf("GetAvailableSpace: Unable to inspect device: '%v'", path)
	}
	
	return int(freeBytes), nil
}