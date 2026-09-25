package script

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func WriteProgress(progress int, installPath string, gameId string) error {
	progressFile, err := os.Create(filepath.Join(installPath, gameId, ".marauder.iprogress"))
	if err != nil {
		return err
	}

	progressFile.Write([]byte(strconv.Itoa(progress)))

	return nil
}

func GetProgress(installPath string, gameId string) (int, error) {
	progressFile, err := os.Open(filepath.Join(installPath, gameId, ".marauder.iprogress"))
	if err != nil {
		return 0, err
	}

	progressBytes, err := io.ReadAll(progressFile)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(progressBytes))
}