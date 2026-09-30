package script

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// There is no pure-go implementation for 'rsync -a' so this is a recreated simpler version
func RsyncA(source string, destination string) error {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("RsyncA: Unable to inspect file: '%v'", source)
	}

	if !sourceInfo.IsDir() {
		return fmt.Errorf("RsyncA: Source is not a directory: '%v'", source)
	}

	err = os.MkdirAll(destination, sourceInfo.Mode().Perm())
	if err != nil {
		return fmt.Errorf("RsyncA: Unable to create destination directory: %v", err)
	}

	entries, err := os.ReadDir(source)
	if err != nil {
		return fmt.Errorf("RsyncA: Unable to list directory: '%v'", source)
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		destinationPath := filepath.Join(destination, entry.Name())

		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("RsyncA: Unable to inspect file: '%v'", sourcePath)
		}

		if info.IsDir() {
			err := RsyncA(sourcePath, destinationPath)
			if err != nil {
				return err
			}

			continue
		}

		in, err := os.Open(sourcePath)
		if err != nil {
			return fmt.Errorf("RsyncA: Unable to open file: '%v'", sourcePath)
		}
		defer in.Close()

		out, err := os.Create(destinationPath)
		if err != nil {
			return fmt.Errorf("RsyncA: Unable to create file: '%v'", destinationPath)
		}
		defer out.Close()

		_, err = io.Copy(out, in)
		if err != nil {
			return fmt.Errorf("RsyncA: Unable to write to file: '%v'", destinationPath)
		}
	}

	return nil
}