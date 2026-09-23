package script

import (
	"errors"
	"strings"

	"github.com/buildkite/shellwords"
)

func ParseScriptVariables(script []string) map[string]string {
	vars := map[string]string{}

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