package script

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
)

// marauderctl remembers your install progress by steps
// If you interrupt an installation you will resume from the step you left
// The progress file gets deleted once the installation is successfully finished

// Writes the progress to the progress file
func WriteProgress(progress int, gameMeta cstructs.GameMeta) error {
	progressFile, err := os.Create(gameMeta.GetProgressFile())
	if err != nil {
		return fmt.Errorf("WriteProgress: Unable to open file: '%v'", gameMeta.GetProgressFile())
	}

	_, err = progressFile.Write([]byte(strconv.Itoa(progress)))
	if err != nil {
		return fmt.Errorf("WriteProgress: Unable to write to file: '%v'", gameMeta.GetProgressFile())
	}

	return nil
}

// Returns the currenta progress
func GetProgress(gameMeta cstructs.GameMeta) (int, error) {
	progressFile, err := os.Open(gameMeta.GetProgressFile())
	if err != nil {
		return 0, fmt.Errorf("GetProgress: Unable to open file: '%v'", gameMeta.GetProgressFile())
	}

	progressBytes, err := io.ReadAll(progressFile)
	if err != nil {
		return 0, fmt.Errorf("GetProgress: Unable to read file: '%v'", gameMeta.GetProgressFile())
	}

	return strconv.Atoi(string(progressBytes))
}