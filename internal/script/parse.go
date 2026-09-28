package script

import (
	"encoding/json"
	"io"
	"os"
	"strings"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"github.com/buildkite/shellwords"
)

func ParseGlobalScriptVariables(fullScript []string) map[string]string {
	vars := map[string]string{}

	for _, line := range fullScript {
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

func ParseLocalScriptVariables(gameMeta cstructs.GameMeta) (map[string]string, error) {
	marauderEnvFile, err := os.Open(gameMeta.GetEnvFile())
	if err != nil {
		return nil, err
	}

	env, err := io.ReadAll(marauderEnvFile)
	if err != nil {
		return nil, err
	}

	var vars map[string]string

	err = json.Unmarshal(env, &vars)
	if err != nil {
		return nil, err
	}

	return vars, nil
}

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