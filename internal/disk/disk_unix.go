//go:build unix

package disk

import "golang.org/x/sys/unix"

func GetFreeSpace(path string) (int, error) {
	var stat unix.Statfs_t

	err := unix.Statfs(path, &stat)
	if err != nil {
		return 0, err
	}

	return int(stat.Bavail) * int(stat.Bsize), nil
}