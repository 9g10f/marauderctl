package script

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"github.com/buildkite/shellwords"
)

// Returns only the global environment variables
func ParseGlobalScriptVariables(fullScript []string) map[string]string {
	vars := map[string]string{}

	for _, line := range fullScript {
		// We can ignore this error because it is always checked by ValidateScript beforehand
		cmd, _ := shellwords.Split(line)

		if len(cmd) > 0 {
			if cmd[0] == "set" {
				definition := strings.Split(strings.Join(cmd[1:], ""), "=")

				variableName := definition[0]
				variableValue := strings.Join(definition[1:], "=")

				vars[variableName] = variableValue
			}
		}

		if IsVersionDelimiter(line) {
			break
		}
	}
	
	return vars
}

// Returns all environment variables (global and versioned)
func ParseScriptVariables(fullScript []string, gameMeta cstructs.GameMeta) map[string]string {
	var script []string

	vars := map[string]string{}

	gameVersion := gameMeta.GetVersion()
	if gameVersion == "latest" {
		gameVersion = ParseGameLatestVersion(fullScript)
	}

	// Process global variables and create versioned script
	stop := false
	for linen, line := range fullScript {
		// We can ignore this error because it is always checked by ValidateScript beforehand
		cmd, _ := shellwords.Split(line)

		if len(cmd) > 0 {
			if cmd[0] == "set" {
				definition := strings.Split(strings.Join(cmd[1:], ""), "=")

				variableName := definition[0]
				variableValue := strings.Join(definition[1:], "=")

				vars[variableName] = variableValue
			}
		}

		if IsVersionDelimiter(line) && line[1:] == gameVersion {
			if !stop {
				for _, subLine := range fullScript[linen + 1:] {
					if IsVersionDelimiter(subLine) {
						break
					}

					script = append(script, subLine)
				}

				stop = true
			}
		}
	}

	// Process versioned variables with a higher priority
	for _, line := range script {
		// We can ignore this error because it is always checked by ValidateScript beforehand
		cmd, _ := shellwords.Split(line)

		if len(cmd) > 0 {
			if cmd[0] == "set" {
				definition := strings.Split(strings.Join(cmd[1:], ""), "=")

				variableName := definition[0]
				variableValue := strings.Join(definition[1:], "=")

				vars[variableName] = variableValue
			}
		}
	}
	
	return vars
}

// Returns the local environment variables of a specified game
// The local environment variables are the mix of the globals and the versioned ones (prioritized)
func ParseLocalScriptVariables(gameMeta cstructs.GameMeta) (map[string]string, error) {
	marauderEnvFile, err := os.Open(gameMeta.GetEnvFile())
	if err != nil {
		return nil, fmt.Errorf("ParseLocalScriptVariables: Unable to open environment file: '%v'", gameMeta.GetEnvFile())
	}

	env, err := io.ReadAll(marauderEnvFile)
	if err != nil {
		return nil, fmt.Errorf("ParseLocalScriptVariables: Unable to read environment file: '%v'", gameMeta.GetEnvFile())
	}

	var vars map[string]string

	err = json.Unmarshal(env, &vars)
	if err != nil {
		return nil, fmt.Errorf("ParseLocalScriptVariables: Unable to parse environment variables")
	}

	return vars, nil
}

// Returns all game versions
func ParseGameVersions(script []string) []string {
	var versions []string

	for _, line := range script {
		if IsVersionDelimiter(line) {
			if len(line) == 1 { // In this stage we haven't checked for blank versions yet
				versions = append(versions, line[1:])
			}
		}
	}

	return versions
}

// Returns the latest version of a certain game
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