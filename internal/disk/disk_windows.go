//go:build windows

package disk

import "golang.org/x/sys/windows"

func GetFreeSpace(path string) (int, error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}

	var freeBytes uint64
	var totalBytes uint64
	var totalFreeBytes uint64

	err = windows.GetDiskFreeSpaceEx(p, &freeBytes, &totalBytes, &totalFreeBytes)
	
	return int(freeBytes), err
}