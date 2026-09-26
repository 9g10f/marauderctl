package script

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

// There is no pure-go implementation for 'rsync -a' so this is a recreated simpler version
func RsyncA(source string, destination string) error {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	if !sourceInfo.IsDir() {
		return errors.New("Source is not a directory")
	}

	err = os.MkdirAll(destination, sourceInfo.Mode().Perm())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		destinationPath := filepath.Join(destination, entry.Name())

		info, err := entry.Info()
		if err != nil {
			return err
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
			return err
		}

		out, err := os.Create(destinationPath)
		if err != nil {
			in.Close()
			return err
		}

		_, err = io.Copy(out, in)
		if err != nil {
			in.Close()
			out.Close()
			return err
		}

		err = in.Close()
		if err != nil {
			return err
		}

		err = out.Close()
		if err != nil {
			return err
		}
	}

	return nil
}