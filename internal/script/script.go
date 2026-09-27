package script

import (
	"errors"
	"strings"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"github.com/buildkite/shellwords"
)

func ReformatScript(script []string) []string {
	newScript := []string{}

	for _, line := range script {
		newScript = append(newScript, strings.TrimSpace(line))
	}

	return newScript
}

func IsVersionDelimiter(line string) bool {
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

func CountVersions(script []string) int {
	count := 0

	for _, line := range script {
		if IsVersionDelimiter(line) {
			count++
		}
	}

	return count
}

func GetVersionedScript(gameMeta cstructs.GameMeta) ([]string, error) {
	var script []string

	fullScript, err := GetScript(gameMeta)
	if err != nil {
		return nil, err
	}

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
		return nil, errors.New("Game version not found")
	} else {
		return script, nil
	}
}

func GetVersionedScriptFromFullScript(fullScript []string, gameMeta cstructs.GameMeta) ([]string, error) {
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
		return nil, errors.New("Game version not found")
	} else {
		return script, nil
	}
}

func GetProcessedFilePath(path string, gameMeta cstructs.GameMeta) string {
	return strings.ReplaceAll(path, "$path", gameMeta.GetDirectory())
}