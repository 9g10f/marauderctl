package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/anacrolix/torrent"
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

func ValidateScript(script []string) error {
	if CountVersions(script) == 0 {
		return errors.New("Game install script is invalid: (-) No versions found")
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

					if !strings.HasPrefix(downloadURL, "http:") && !strings.HasPrefix(downloadURL, "https:") && !strings.HasPrefix(downloadURL, "magnet:") {
						return fmt.Errorf("Game install script is invalid: (%v) Unsupported download URL: '%v'", linen + 1, downloadURL)
					}

					_, err := url.Parse(downloadURL)
					if err != nil {
						return fmt.Errorf("Game install script is invalid: (%v) Invalid download URL: '%v' | error=%v", linen + 1, downloadURL, err)
					}
				}
			case "unzip":
			case "rm":
			case "rsynca":
				for i, rawFilepaths := range cmd {
					if i == 0 {
						continue
					}

					if strings.Count(rawFilepaths, "|") != 1 {
						return fmt.Errorf("Game install script is invalid: (%v) Invalid argument for rsynca: '%v', exactly two parts (separated by a '|' character) are required", rawFilepaths, linen + 1)
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

func DownloadHTTP(downloadURL string, installPath string) error {
	r, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	u, _ := url.Parse(downloadURL)
	filename := path.Base(u.Path)

	file, err := os.Create(installPath + string(filepath.Separator) + filename)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, r.Body) // TODO: Get live progress
	if err != nil {
		file.Close()
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

func DownloadTorrentMagnet(downloadURL string, installPath string) error {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = installPath

	client, err := torrent.NewClient(cfg)
	if err != nil {
		return err
	}

	t, err := client.AddMagnet(downloadURL)
	if err != nil {
		client.Close()
		return err
	}

	<-t.GotInfo()
	t.DownloadAll()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	var last int64 = t.BytesCompleted()

	speedsN := 0
	var speedSum int64 = 0
	for range ticker.C {
		info := t.Info()

		total := info.TotalLength()
		completed := t.BytesCompleted()
		percent := float64(completed) / float64(total) * 100
		speed := completed - last
		speedsN++
		speedSum += speed
		last = completed

		avgSpeed := speedSum / int64(speedsN)

		var eta int64 = 0
		if avgSpeed > 0 {
			eta = (total - completed) / avgSpeed
		}

		hours := eta / 3600
		minutes := (eta % 3600) / 60
		seconds := eta % 60

		fmt.Printf("\r\033[2KDownloading game files ...   %.2f%% (%d / %d bytes) @ %.2f MiB/s ETA %d:%02d:%02d", percent, completed, total, float64(speed) / 1024 / 1024, hours, minutes, seconds)

		if completed == total {
			break
		}
	}

	fmt.Println()

	client.WaitAll()

	// Wait one second before dropping the torrent
	// Dropping the torrent stops seeding
	time.Sleep(1 * time.Second)

	t.Drop()

	errs := client.Close()
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	err = os.Remove(filepath.Join(installPath, ".torrent.db"))
	if err != nil {
		return err
	}

	return nil
}

// ERROR_SHARING_VIOLATION is not exported by the standard syscall package on Windows
// Its value (32) is fixed by the Windows API
const ERROR_SHARING_VIOLATION syscall.Errno = 32

func RunScript(script []string, force bool, installPath string, resumeline int) error {
	for linen, line := range script {
		if linen + 1 < resumeline {
			continue
		}

		cmd, _ := shellwords.Split(line)
		
		if len(cmd) > 0 { // Skip blank lines
			fmt.Println(line) // REMOVE
			// fmt.Scanln() // REMOVE

			switch cmd[0] {
			case "set":
				// Variables are handled by ParseScriptVariables
				continue
			case "download":
				// The download command can take various arguments (URLs), of clearnet, BitTorrent and in the future Tor files
				// It downloads all files onto 'installPath' with the name on the URL
				for i, downloadURL := range cmd {
					if i == 0 {
						continue
					}

					if strings.HasPrefix(downloadURL, "http:") || strings.HasPrefix(downloadURL, "https:") {
						err := DownloadHTTP(downloadURL, installPath)
						if err != nil {
							return err
						}
					} else {
						err := DownloadTorrentMagnet(downloadURL, installPath)
						if err != nil {
							return err
						}
					}
				}
			case "unzip":
				// The unzip command can take various arguments (paths) and uses 7-Zip (a dependency) to unzip all files
				// The files are unziped directly to their current directory
				for i, rawFilepath := range cmd {
					if i == 0 {
						continue
					}

					trueFilepath := GetProcessedFilePath(rawFilepath, installPath)

					cmd := exec.Command(
						"7z",
						"x",
						trueFilepath,
						"-ponline-fix.me", // TODO: Needs to be customizable in the future
						"-y",
					)

					cmd.Dir = filepath.Dir(trueFilepath)

					err := cmd.Run()
					if err != nil {
						return err
					}
				}
			case "rm":
				// The rm command can take various arguments (paths) and removes all files/directories
				for i, rawFilepath := range cmd {
					if i == 0 {
						continue
					}

					trueFilepath := GetProcessedFilePath(rawFilepath, installPath)

					var err error
					success := false
					for attempts := 0; attempts < 40; attempts++ {
						err = os.RemoveAll(trueFilepath)
						if err == nil {
							success = true
							break
						}

						pathError, ok := err.(*os.PathError)
						if ok {
							sysError, ok := pathError.Err.(syscall.Errno)
							if ok {
								// Linux doesn't have an equivalent to Windows sharing violation so it won't throw an error in that situation
								// Checking for it everytime isn't harmfull on Linux machines because the error that shares the same error number, EPIPE (Broken pipe), can't happen in an unlink or rmdir syscall
								if sysError == ERROR_SHARING_VIOLATION {
									// If the file(s) we are trying to delete is/are locked, retry 25 times with 250 millisecond intervals before finally crashing
									time.Sleep(250 * time.Millisecond)
									continue
								}
							}
						}

						return err
					}

					if !success {
						return err
					}
				}
			case "rsynca":
				// The rsynca command can take various pairs of arguments (path|path) separated by the '|' character and merges two folders overwriting the second one (destination)
				// It runs very similar to 'rsync -a', but we decided not to include rsync as a dependency for Windows users
				for i, rawFilepaths := range cmd {
					if i == 0 {
						continue
					}

					rawFilepathSource := strings.Split(rawFilepaths, "|")[0]
					rawFilepathDestination := strings.Split(rawFilepaths, "|")[1]

					filepathSource := GetProcessedFilePath(rawFilepathSource, installPath)
					filepathDestination := GetProcessedFilePath(rawFilepathDestination, installPath)

					err := RsyncA(filepathSource, filepathDestination)
					if err != nil {
						return err
					}
				}
			case "mv":
				// The mv command can take various pairs of arguments (path|path) separated by the '|' character and moves a file or a directory
				// Mostly identical to the Unix mv command
				// Supports glob
				for i, rawFilepaths := range cmd {
					if i == 0 {
						continue
					}

					rawFilepathSource := strings.Split(rawFilepaths, "|")[0]
					rawFilepathDestination := strings.Split(rawFilepaths, "|")[1]

					filepathSource := GetProcessedFilePath(rawFilepathSource, installPath)
					filepathDestination := GetProcessedFilePath(rawFilepathDestination, installPath)

					filepathSourceGlob, _ := filepath.Glob(filepathSource)
					if filepathSourceGlob == nil {
						return fmt.Errorf("File(s) not found: '%v'", filepathSource)
					}

					filepathDestinationGlob, _ := filepath.Glob(filepathDestination)
					if filepathDestinationGlob == nil {
						return fmt.Errorf("File(s) not found: '%v'", filepathDestination)
					}

					if len(filepathDestinationGlob) != 1 {
						return fmt.Errorf("Too many destinations: '%v'", filepathDestination)
					}

					sources := filepathSourceGlob
					destination := filepathDestinationGlob[0]

					destinationInfo, err := os.Stat(destination)
					if os.IsNotExist(err) {
						if err := os.Rename(sources[0], destination); err != nil {
							return err
						}

						continue
					} else if err != nil {
						return err
					}

					if !destinationInfo.IsDir() {
						return fmt.Errorf("Destination is not a directory: '%v'", destination)
					}

					for _, source := range sources {
						target := filepath.Join(destination, filepath.Base(source))

						err = os.Rename(source, target)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}

	return nil
}