package script

import (
	"errors"
	"strings"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"github.com/buildkite/shellwords"
)

// Removes leading and trailing spaces which can break RunScript
func ReformatScript(script []string) []string {
	newScript := []string{}

	for _, line := range script {
		newScript = append(newScript, strings.TrimSpace(line))
	}

	return newScript
}

// Returns if a line defines the start of a versioned script (@<game-version>)
func IsVersionDelimiter(line string) bool {
	// We can ignore this error because it is always checked by ValidateScript beforehand
	cmd, _ := shellwords.Split(line)

	if len(cmd) > 0 {
		if strings.HasPrefix(cmd[0], "@") {
			return true
		} else {
			return false
		}
	}

	return false
}

// Returns the number of versions in a script
func CountVersions(script []string) int {
	count := 0

	for _, line := range script {
		if IsVersionDelimiter(line) {
			count++
		}
	}

	return count
}

// Returns the cut/versioned part from a script
func GetVersionedScript(fullScript []string, gameMeta cstructs.GameMeta) ([]string, error) {
	var script []string

	gameVersion := gameMeta.GetVersion()
	if gameVersion == "latest" {
		gameVersion = ParseGameLatestVersion(fullScript)
	}

	found := false
	for linen, line := range fullScript {
		if IsVersionDelimiter(line) && line[1:] == gameVersion {
			for _, subLine := range fullScript[linen + 1:] {
				if IsVersionDelimiter(subLine) {
					break
				}

				script = append(script, subLine)
			}

			found = true
			break
		}
	}

	if !found {
		return nil, errors.New("GetVersionedScript: Game version not found")
	} else {
		return script, nil
	}
}

// Install scripts don't use regular paths, they use relative paths where '$path' is replace with the game's install directory
func GetProcessedFilePath(path string, gameMeta cstructs.GameMeta) string {
	return strings.ReplaceAll(path, "$path", gameMeta.GetDirectory())
}