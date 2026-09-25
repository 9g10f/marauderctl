package script

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/buildkite/shellwords"
)

func ParseScriptVariables(game string, fullScript []string) map[string]string {
	var script []string

	vars := map[string]string{}

	gameVersion := GetGameVersion(game)
	if gameVersion == "latest" {
		gameVersion = ParseGameLatestVersion(fullScript)
	}

	// Process global variables and create versioned script
	stop := false
	for linen, line := range fullScript {
		cmd, _ := shellwords.Split(line)

		if len(cmd) > 0 {
			if cmd[0] == "set" {
				defenition := strings.Split(strings.Join(cmd[1:], ""), "=")

				variableName := defenition[0]
				variableValue := strings.Join(defenition[1:], "=")

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
				defenition := strings.Split(strings.Join(cmd[1:], ""), "=")

				variableName := defenition[0]
				variableValue := strings.Join(defenition[1:], "=")

				vars[variableName] = variableValue
			}
		}
	}
	
	return vars
}

func ParseLocalScriptVariables(game string, installPath string, server string) (map[string]string, error) {
	marauderEnvFile, err := os.Open(filepath.Join(GetGameFolder(installPath, server, game), ".marauder.env"))
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