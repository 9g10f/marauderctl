package main

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
)

func CountVersions(script []string) int {
	count := 0

	for _, line := range script {
		if strings.HasPrefix(line, "@") {
			count++
		}
	}

	return count
}

func GetInstallScript(game string, server string) ([]string, error) {
	var script []string
	var fullScript []string
	var gameId string
	var gameVersion string

	if strings.Count(game, "@") > 1 {
		return nil, errors.New("Invalid game field format")
	}

	if !strings.Contains(game, "@") {
		gameId = game
		gameVersion = "latest"
	} else {
		gameId = strings.Split(game, "@")[0]
		gameVersion = strings.Split(game, "@")[1]
	}

	if before, ok := strings.CutPrefix(server, "file://"); ok {
		data, err := os.ReadFile(before + gameId)
		if err != nil {
			return nil, err
		}

		fullScript = strings.Split(string(data), "\n")
	} else {
		r, err := http.Get(server + gameId)
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()

		data, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		fullScript = strings.Split(string(data), "\n")
	}

	nGameVersions := CountVersions(fullScript)
	if nGameVersions == 0 {
		return nil, errors.New("Corrupted game install script, no versions found")
	}

	if gameVersion == "latest" {
		currentVersionCount := 0
		for idx, line := range fullScript {
			if strings.HasPrefix(line, "@") {
				currentVersionCount++
			}

			if currentVersionCount == nGameVersions {
				for _, subLine := range fullScript[idx + 1:] {
					if strings.HasPrefix(subLine, "@") {
						break
					}

					script = append(script, subLine)
				}

				break
			}
		}
	} else {
		found := false
		for idx, line := range fullScript {
			if strings.HasPrefix(line, "@") && line[1:] == gameVersion {
				for _, subLine := range fullScript[idx + 1:] {
					if strings.HasPrefix(subLine, "@") {
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
		}
	}

	return script, nil
}