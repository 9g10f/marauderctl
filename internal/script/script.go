package script

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/buildkite/shellwords"
)

// There is no pure-go implementation for 'rsync -a' so this is a recreated simpler version
func RsyncA(source string, destination string) error {
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	if !sourceInfo.IsDir() {
		return errors.New("Source is not a directory")
	}

	err = os.MkdirAll(destination, sourceInfo.Mode().Perm())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		destinationPath := filepath.Join(destination, entry.Name())

		info, err := entry.Info()
		if err != nil {
			return err
		}

		if info.IsDir() {
			err := RsyncA(sourcePath, destinationPath)
			if err != nil {
				return err
			}

			continue
		}

		in, err := os.Open(sourcePath)
		if err != nil {
			return err
		}

		out, err := os.Create(destinationPath)
		if err != nil {
			in.Close()
			return err
		}

		_, err = io.Copy(out, in)
		if err != nil {
			in.Close()
			out.Close()
			return err
		}

		err = in.Close()
		if err != nil {
			return err
		}

		err = out.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

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

func GetProcessedFilePath(path string, installPath string) string {
	return strings.ReplaceAll(path, "$path", installPath)
}

func ValidateScript(script []string, game string) error {
	if CountVersions(script) == 0 {
		return errors.New("Game install script is invalid: (-) No versions found")
	}

	gameId := GetGameId(game)

	for _, version := range ParseGameVersions(script) {
		vars := ParseScriptVariables(gameId + "@" + version, script)

		_, ok := vars["dir"]
		if !ok {
			return fmt.Errorf("Game install script is invalid: (-) Version %v doesn't have a directory variable set", version)
		}
	}

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
					trueFilepath := GetProcessedFilePath(rawFilepath, "./")

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
					processedFilepath := GetProcessedFilePath(rawFilepath, "./")

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
					filepathSource := GetProcessedFilePath(rawFilepathSource, "./")
					filepathDestination := GetProcessedFilePath(rawFilepathDestination, "./")

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
					filepathSource := GetProcessedFilePath(rawFilepathSource, "./")
					filepathDestination := GetProcessedFilePath(rawFilepathDestination, "./")

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