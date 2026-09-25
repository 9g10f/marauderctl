package script

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"slices"
	"strings"

	"github.com/buildkite/shellwords"
)

func ValidateScript(script []string, game string) error {
	if CountVersions(script) == 0 {
		return errors.New("Game install script is invalid: (-) No versions found")
	}

	gameId := GetGameId(game)

	for linen, line := range script {
		cmd, err := shellwords.Split(line)
		if len(cmd) > 0 {
			if err != nil {
				return fmt.Errorf("Game install script is invalid: (%v) Line can't be parsed | error=%v", linen + 1, err)
			}

			if cmd[0][0] == '@' {
				if cmd[0] == "@" {
					return fmt.Errorf("Game install script is invalid: (%v) Version delimiter has no name", linen + 1)
				}

				continue
			}

			switch cmd[0] {
			case "set":
				// The 'set' command requires scripts to include whitespaces before and after the "=" character. If the script has the "=" character between the variable name and the variable value wihtout any whitespaces it will be parsed as being a single part (the name).
				if !slices.Contains(cmd, "=") {
					return fmt.Errorf("Game install script is invalid: (%v) Variable defenition does not contain a name and a value (whitespaces are required before and after the '=' character)", linen + 1)
				}
			case "download":
				for i, downloadURL := range cmd {
					if i == 0 {
						continue
					}
					
					_, err := url.Parse(downloadURL)
					if err != nil {
						return fmt.Errorf("Game install script is invalid: (%v) Invalid download URL: '%v' | error=%v", linen + 1, downloadURL, err)
					}
				}
			case "unzip":
				for i, rawFilepath := range cmd {
					if i == 0 {
						continue
					}

					// Parse dummy install paths just to test glob syntax
					trueFilepath := GetProcessedFilePath(rawFilepath, "./", gameId)

					_, err := filepath.Glob(trueFilepath)
					if err != nil {
						return fmt.Errorf("Game install script is invalid: (%v) Invalid glob path argument for unzip: '%v' | error=%v", linen + 1, rawFilepath, err)
					}
				}
			case "rm":
				for i, rawFilepath := range cmd {
					if i == 0 {
						continue
					}

					// Parse dummy install paths just to test glob syntax
					processedFilepath := GetProcessedFilePath(rawFilepath, "./", gameId)

					_, err := filepath.Glob(processedFilepath)
					if err != nil {
						return fmt.Errorf("Game install script is invalid: (%v) Invalid glob path argument for rm: '%v' | error=%v", linen + 1, rawFilepath, err)
					}
				}
			case "rsynca":
				for i, rawFilepaths := range cmd {
					if i == 0 {
						continue
					}

					if strings.Count(rawFilepaths, "|") != 1 {
						return fmt.Errorf("Game install script is invalid: (%v) Invalid argument for rsynca: '%v', exactly two parts (separated by a '|' character) are required", rawFilepaths, linen + 1)
					}

					rawFilepathSource := strings.Split(rawFilepaths, "|")[0]
					rawFilepathDestination := strings.Split(rawFilepaths, "|")[1]

					// Parse dummy install paths just to test glob syntax
					filepathSource := GetProcessedFilePath(rawFilepathSource, "./", gameId)
					filepathDestination := GetProcessedFilePath(rawFilepathDestination, "./", gameId)

					_, err := filepath.Glob(filepathSource)
					if err != nil {
						return fmt.Errorf("Game install script is invalid: (%v) Invalid glob path argument for rsynca: '%v' | error=%v", linen + 1, filepathSource, err)
					}

					_, err = filepath.Glob(filepathDestination)
					if err != nil {
						return fmt.Errorf("Game install script is invalid: (%v) Invalid glob path argument for rsynca: '%v' | error=%v", linen + 1, filepathDestination, err)
					}
				}
			case "mv":
				for i, rawFilepaths := range cmd {
					if i == 0 {
						continue
					}

					if strings.Count(rawFilepaths, "|") != 1 {
						return fmt.Errorf("Game install script is invalid: (%v) Invalid argument for mv: '%v', exactly two parts (separated by a '|' character) are required", rawFilepaths, linen + 1)
					}

					rawFilepathSource := strings.Split(rawFilepaths, "|")[0]
					rawFilepathDestination := strings.Split(rawFilepaths, "|")[1]

					// Parse dummy install paths just to test glob syntax
					filepathSource := GetProcessedFilePath(rawFilepathSource, "./", gameId)
					filepathDestination := GetProcessedFilePath(rawFilepathDestination, "./", gameId)

					_, err := filepath.Glob(filepathSource)
					if err != nil {
						return fmt.Errorf("Game install script is invalid: (%v) Invalid glob path argument for mv: '%v' | error=%v", linen + 1, filepathSource, err)
					}

					_, err = filepath.Glob(filepathDestination)
					if err != nil {
						return fmt.Errorf("Game install script is invalid: (%v) Invalid glob path argument for mv: '%v' | error=%v", linen + 1, filepathDestination, err)
					}
				}
			default:
				return fmt.Errorf("Game install script is invalid: (%v) Invalid command", linen + 1)
			}
		}
	}

	return nil
}