package disk

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"codeberg.org/9g10f/marauderctl/internal/script"
)

// Used to check for available space on the game's device
// Returns an error only if the device doesn't contain enough space (game size + 10 MB)
func VerifyAvailableSize(gameMeta cstructs.GameMeta) error {
	freeSpace, err := GetAvailableSpace(root(gameMeta.InstallPath))
	if err != nil {
		return err
	}

	fullScript, err := script.GetScript(gameMeta)
	if err != nil {
		return err
	}

	vars := script.ParseScriptVariables(fullScript, gameMeta)

	maxSize, ok := vars["max-size"]
	if !ok {
		return nil
	}

	maxSizeI, err := strconv.Atoi(maxSize)
	if err != nil {
		return fmt.Errorf("VerifyAvailableSize: Variable maxSizeI: '%v' can not be converted into an integer", maxSize)
	}

	// Give +10 MB of room
	if freeSpace < (maxSizeI + int(math.Pow(10, 7))) {
		return fmt.Errorf("VerifyAvailableSize: No space left on device: '%v'", root(gameMeta.InstallPath))
	}

	return nil
}

// Returns the current device's root from a path
func root(p string) string {
	p = filepath.Clean(p)

	for {
		_, err := os.Stat(p)
		if err == nil {
			return p
		}

		parent := filepath.Dir(p)
		if parent == p {
			return p
		}
		
		p = parent
	}
}