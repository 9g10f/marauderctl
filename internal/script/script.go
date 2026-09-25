package script

import (
	"errors"
	"io"
	"os"
	"path/filepath"
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

func GetVersionedScriptFromFullScript(game string, fullScript []string) ([]string, error) {
	var script []string

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