package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/buildkite/shellwords"
)

func ReformatScript(script []string) []string {
	newScript := []string{}

	for _, line := range script {
		if strings.TrimSpace(line) != "" { // Remove empty lines
			newScript = append(newScript, strings.TrimSpace(line))
		}
	}

	return newScript
}

func IsVersionDelimiter(line string) bool {
	cmd, _ := shellwords.Split(line)

	if strings.HasPrefix(cmd[0], "@") {
		return true
	} else {
		return false
	}
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

func GetGameVersion(game string) string {
	var version string

	if !strings.Contains(game, "@") {
		version = "latest"
	} else {
		version = strings.Split(game, "@")[1]
	}

	return version
}

func GetGameId(game string) string {
	var gameId string

	if !strings.Contains(game, "@") {
		gameId = game
	} else {
		gameId = strings.Split(game, "@")[0]
	}

	return gameId
}

func ValidateScript(script []string) error {
	if CountVersions(script) == 0 {
		return errors.New("Game install script is invalid: (-) No versions found")
	}

	for linen, line := range script {
		cmd, err := shellwords.Split(line)
		if err != nil {
			return fmt.Errorf("Game install script is invalid: (%v) Line can't be parsed", linen)
		}

		if cmd[0][0] == '@' {
			if cmd[0] == "@" {
				return fmt.Errorf("Game install script is invalid: (%v) Version delimiter has no name", linen)
			}
		}

		switch cmd[0] {
		case "set":
			// The 'set' command requires scripts to include whitespaces before and after the "=" character. If the script has the "=" character between the variable name and the variable value wihtout any whitespaces it will be parsed as being a single part: the name.
			if !slices.Contains(cmd, "=") {
				return fmt.Errorf("Game install script is invalid: (%v) Variable defenition does not contain a name and a value (whitespaces are required before and after the \"=\" character)", linen)
			}
		}
	}

	return nil
}

func GetVersionedScript(game string, server string) ([]string, error) {
	var script []string

	fullScript, err := GetScript(game, server)
	if err != nil {
		return nil, err
	}

	gameVersion := GetGameVersion(game)
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

func ParseScriptVariables(script []string) map[string]string {
	vars := map[string]string{}

	for _, line := range script {
		cmd, _ := shellwords.Split(line)

		if cmd[0] == "set" {
			defenition := strings.Split(strings.Join(cmd[1:], ""), "=")

			variableName := defenition[0]
			variableValue := strings.Join(defenition[1:], "=")

			vars[variableName] = variableValue
		}
	}
	
	return vars
}

func ParseGameName(script []string) (string, error) {
	vars := ParseScriptVariables(script)

	name, ok := vars["name"]
	if ok {
		return name, nil
	} else {
		return "", errors.New("Game script variable: 'name' was never set")
	}
}

func ParseGameDir(script []string) (string, error) {
	vars := ParseScriptVariables(script)

	name, ok := vars["dir"]
	if ok {
		return name, nil
	} else {
		return "", errors.New("Game script variable: 'dir' was never set")
	}
}

func ParseGameExe(script []string) (string, error) {
	vars := ParseScriptVariables(script)

	name, ok := vars["exe"]
	if ok {
		return name, nil
	} else {
		return "", errors.New("Game script variable: 'exe' was never set")
	}
}

func ParseGameMaxSize(script []string) (string, error) {
	vars := ParseScriptVariables(script)

	name, ok := vars["max-size"]
	if ok {
		return name, nil
	} else {
		return "", errors.New("Game script variable: 'max-size' was never set")
	}
}

func ParseGameSize(script []string) (string, error) {
	vars := ParseScriptVariables(script)

	name, ok := vars["size"]
	if ok {
		return name, nil
	} else {
		return "", errors.New("Game script variable: 'size' was never set")
	}
}

func ParseGameLatestVersion(script []string) string {
	nGameVersions := CountVersions(script)

	currentVersionCount := 0
	for _, line := range script {
		if IsVersionDelimiter(line) {
			currentVersionCount++
		}

		if currentVersionCount == nGameVersions {
			return line[1:]
		}
	}

	return ""
}

func RunScript(script []string, force bool) error {
	for linen := range script {
		fmt.Printf("\r%v/%v", linen, len(script))
	}

	fmt.Printf("\r%v/%v", len(script), len(script))

	return nil
}