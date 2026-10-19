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

// Validates the script's syntax
func ValidateScript(script []string) error {
	if CountVersions(script) == 0 {
		return errors.New("Validate: Game install script is invalid: (-) No versions found")
	}

	for linen, line := range script {
		cmd, err := shellwords.SplitPosix(line)
		if err != nil {
			return fmt.Errorf("Validate: Game install script is invalid: (%v) Line can't be parsed | error=%v", linen + 1, err)
		}

		if len(cmd) > 0 {
			if cmd[0][0] == '@' {
				if cmd[0] == "@" {
					// Version can't be blank
					return fmt.Errorf("Validate: Game install script is invalid: (%v) Version delimiter has no name", linen + 1)
				}

				continue
			}

			if cmd[0][0] == '#' {
				// Allow comments
				continue
			}

			switch cmd[0] {
			case "set":
				if slices.Contains(cmd, "=") {
					if len(cmd) < 4 {
						return fmt.Errorf("Validate: Game install script is invalid: (%v) Variable definition does not contain a name and a value", linen + 1)
					}
				} else {
					if strings.Contains(line, "=") {
						separator := GetSetSeparator(line)
						if separator == -1 {
							return fmt.Errorf("Validate: Game install script is invalid: (%v) Variable definition does not contain a '='", linen + 1)
						}
						
						_, err := shellwords.SplitPosix(line[separator + 1:])
						if err != nil {
							return fmt.Errorf("Validate: Game install script is invalid: (%v), Variable defenition value cannot be split by shellwords: %v", linen + 1, err)
						}
					} else {
						return fmt.Errorf("Validate: Game install script is invalid: (%v) Variable definition does not contain a '='", linen + 1)
					}
				}
			case "download":
				for i, downloadURL := range cmd {
					if i == 0 {
						continue
					}
					
					_, err := url.Parse(downloadURL)
					if err != nil {
						return fmt.Errorf("Validate: Game install script is invalid: (%v) Invalid download URL: '%v' | error=%v", linen + 1, downloadURL, err)
					}
				}
			case "unzip":
				for i, rawFilepath := range cmd {
					if i == 0 {
						continue
					}

					// Test the path for glob syntax errors
					_, err := filepath.Glob(strings.ReplaceAll(rawFilepath, "$path", "."))
					if err != nil {
						return fmt.Errorf("Validate: Game install script is invalid: (%v) Invalid glob path argument for unzip: '%v' | error=%v", linen + 1, rawFilepath, err)
					}
				}
			case "rm":
				for i, rawFilepath := range cmd {
					if i == 0 {
						continue
					}

					// Test the path for glob syntax errors
					_, err := filepath.Glob(strings.ReplaceAll(rawFilepath, "$path", "."))
					if err != nil {
						return fmt.Errorf("Validate: Game install script is invalid: (%v) Invalid glob path argument for rm: '%v' | error=%v", linen + 1, rawFilepath, err)
					}
				}
			case "rsynca":
				for i, rawFilepaths := range cmd {
					if i == 0 {
						continue
					}

					if strings.Count(rawFilepaths, "|") != 1 {
						return fmt.Errorf("Validate: Game install script is invalid: (%v) Invalid argument for rsynca: '%v', exactly two parts (separated by a '|' character) are required", rawFilepaths, linen + 1)
					}

					rawFilepathSource := strings.Split(rawFilepaths, "|")[0]
					rawFilepathDestination := strings.Split(rawFilepaths, "|")[1]

					// Test the path for glob syntax errors
					_, err := filepath.Glob(strings.ReplaceAll(rawFilepathSource, "$path", "."))
					if err != nil {
						return fmt.Errorf("Validate: Game install script is invalid: (%v) Invalid glob path argument for rsynca: '%v' | error=%v", linen + 1, rawFilepathSource, err)
					}

					// Test the path for glob syntax errors
					_, err = filepath.Glob(strings.ReplaceAll(rawFilepathDestination, "$path", "."))
					if err != nil {
						return fmt.Errorf("Validate: Game install script is invalid: (%v) Invalid glob path argument for rsynca: '%v' | error=%v", linen + 1, rawFilepathDestination, err)
					}
				}
			case "mv":
				for i, rawFilepaths := range cmd {
					if i == 0 {
						continue
					}

					if strings.Count(rawFilepaths, "|") != 1 {
						return fmt.Errorf("Validate: Game install script is invalid: (%v) Invalid argument for mv: '%v', exactly two parts (separated by a '|' character) are required", rawFilepaths, linen + 1)
					}

					rawFilepathSource := strings.Split(rawFilepaths, "|")[0]
					rawFilepathDestination := strings.Split(rawFilepaths, "|")[1]

					// Test the path for glob syntax errors
					_, err := filepath.Glob(strings.ReplaceAll(rawFilepathSource, "$path", "."))
					if err != nil {
						return fmt.Errorf("Validate: Game install script is invalid: (%v) Invalid glob path argument for mv: '%v' | error=%v", linen + 1, rawFilepathSource, err)
					}

					// Test the path for glob syntax errors
					_, err = filepath.Glob(strings.ReplaceAll(rawFilepathDestination, "$path", "."))
					if err != nil {
						return fmt.Errorf("Validate: Game install script is invalid: (%v) Invalid glob path argument for mv: '%v' | error=%v", linen + 1, rawFilepathDestination, err)
					}
				}
			default:
				return fmt.Errorf("Validate: Game install script is invalid: (%v) Invalid command", linen + 1)
			}
		}
	}

	return nil
}