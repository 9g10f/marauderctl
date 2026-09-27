package script

import (
	"io"
	"os"
	"path/filepath"
	"strconv"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
)

func WriteProgress(progress int, gameMeta cstructs.GameMeta) error {
	progressFile, err := os.Create(filepath.Join(gameMeta.GetDirectory(), ".marauder.iprogress"))
	if err != nil {
		return err
	}

	progressFile.Write([]byte(strconv.Itoa(progress)))

	return nil
}

func GetProgress(gameMeta cstructs.GameMeta) (int, error) {
	progressFile, err := os.Open(filepath.Join(gameMeta.GetDirectory(), ".marauder.iprogress"))
	if err != nil {
		return 0, err
	}

	progressBytes, err := io.ReadAll(progressFile)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(progressBytes))
}