package script

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func WriteProgress(progress int, installPath string, server string, game string) error {
	progressFile, err := os.Create(filepath.Join(GetGameFolder(installPath, server, game), ".marauder.iprogress"))
	if err != nil {
		return err
	}

	progressFile.Write([]byte(strconv.Itoa(progress)))

	return nil
}

func GetProgress(installPath string, server string, game string) (int, error) {
	progressFile, err := os.Open(filepath.Join(GetGameFolder(installPath, server, game), ".marauder.iprogress"))
	if err != nil {
		return 0, err
	}

	progressBytes, err := io.ReadAll(progressFile)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(progressBytes))
}