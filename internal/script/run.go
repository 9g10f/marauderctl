package script

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"codeberg.org/9g10f/marauderctl/internal/logging"
	"github.com/buildkite/shellwords"
)

// ERROR_SHARING_VIOLATION is not exported by the standard syscall package on Windows
// Its value (32) is fixed by the Windows API
const ERROR_SHARING_VIOLATION syscall.Errno = 32

// Runs the versioned script
func RunScript(script []string, installFlags cstructs.InstallFlags, gameMeta cstructs.GameMeta) error {
	st := time.Now()

	// Log the installation to a file
	logfile, err := logging.GetInstallLogFile(st, gameMeta)
	if err != nil {
		return err
	}
	defer logfile.Close()

	torrentLogger := logging.Logger{
		File: logfile,
		Location: "torrent",
	}

	unzipLogger := logging.Logger{
		File: logfile,
		Location: "7z",
	}

	// Set torrent setting to close mapped files
	_, ok := os.LookupEnv("TORRENT_STORAGE_DEFAULT_FILE_IO")
	if !ok {
		os.Setenv("TORRENT_STORAGE_DEFAULT_FILE_IO", "classic")
	}

	// Resume progress
	progress, err := GetProgress(gameMeta)
	if err != nil {
		progress = 0
	} else {
		installFlags.ResumeLine = progress + 1
	}
	
	err = WriteProgress(progress, gameMeta)
	if err != nil {
		return err
	}

	for linen, line := range script {
		if linen + 1 < installFlags.ResumeLine {
			continue
		}

		command, err := shellwords.Split(line)
		if err != nil {
			return fmt.Errorf("RunScript: Unable to split command: '%v'", command)
		}
		
		if len(command) > 0 { // Skip blank lines
			// Allow comments
			if command[0][0] == '#' {
				progress++

				err := WriteProgress(progress, gameMeta)
				if err != nil {
					return err
				}

				continue
			}

			switch command[0] {
			case "set":
				// Variables are handled by ParseScriptVariables
				progress++

				err := WriteProgress(progress, gameMeta)
				if err != nil {
					return err
				}

				continue
			case "download":
				// The download command can take various arguments (URLs), of clearnet, BitTorrent and in the future Tor files
				// It downloads all files onto 'installPath' with the name on the URL
				for i, downloadURL := range command {
					if i == 0 {
						continue
					}

					if installFlags.OutputStyle == "default" {
						fmt.Printf("\r\033[2KDownloading game files ... \t0%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00")
					} else if installFlags.OutputStyle == "json" {
						fmt.Printf("\r\033[2K{\"task\": \"Downloading game files\", \"details\": \"Downloading %v\", \"progress\": 0, \"eta\": 0}", downloadURL)
					}

					if strings.HasPrefix(downloadURL, "magnet:") {
						err := DownloadTorrentMagnet(downloadURL, installFlags.OutputStyle, torrentLogger, gameMeta)
						if err != nil {
							return err
						}
					} else {
						err := Download(downloadURL, gameMeta)
						if err != nil {
							return err
						}
					}

					if installFlags.OutputStyle == "default" {
						fmt.Printf("\r\033[2KDownloading game files ... \t100%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
					} else if installFlags.OutputStyle == "json" {
						fmt.Printf("\r\033[2K{\"task\": \"Downloading game files\", \"details\": \"Downloading %v\", \"progress\": 100, \"eta\": 0}\n", downloadURL)
					}
				}
			case "unzip":
				// The unzip command can take various arguments (paths) and uses 7-Zip (a dependency) to unzip all files
				// The files are unziped directly to their current directory
				for i, rawFilepath := range command {
					if i == 0 {
						continue
					}

					trueFilepath := GetProcessedFilePath(rawFilepath, gameMeta)

					trueFilepathGlob, err := filepath.Glob(trueFilepath)
					if err != nil {
						return fmt.Errorf("RunScript: unzip: Invalid glob syntax: '%v'", trueFilepath)
					}

					if trueFilepathGlob == nil {
						return fmt.Errorf("RunScript: unzip: No matches for glob: '%v'", trueFilepath)
					}

					for _, source := range trueFilepathGlob {
						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KUnzipping game files ... \t0%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Unzipping game files\", \"details\": \"Unzipping %v\", \"progress\": 0, \"eta\": 0}", source)
						}

						command := exec.Command(
							"7z",
							"x",
							source,
							"-ponline-fix.me",
							"-y",
						)

						command.Dir = filepath.Dir(source)
						command.Stdout = unzipLogger
						command.Stderr = unzipLogger

						err := command.Run()
						if err != nil {
							exitError, ok := err.(*exec.ExitError)
							if !ok {
								return fmt.Errorf("RunScript: unzip: Error while running 7z: %v", err)
							}

							code := exitError.ExitCode()

							// We don't throw an error when the exit code is 2 because that's the exit code 7z throws when even tho it still unzipped the files there were some warnings (e.g. Unsupported Method)
							if code != 2 {
								return fmt.Errorf("RunScript: unzip: Error while running 7z: %v", err)
							}
						}

						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KUnzipping game files ... \t100%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Unzipping game files\", \"details\": \"Unzipping %v\", \"progress\": 100, \"eta\": 0}\n", source)
						}
					}
				}
			case "rm":
				// The rm command can take various arguments (paths) and removes all files/directories
				for i, rawFilepath := range command {
					if i == 0 {
						continue
					}

					trueFilepath := GetProcessedFilePath(rawFilepath, gameMeta)

					trueFilepathGlob, err := filepath.Glob(trueFilepath)
					if err != nil {
						return fmt.Errorf("RunScript: rm: Invalid glob syntax: '%v'", trueFilepath)
					}

					if trueFilepathGlob == nil {
						return fmt.Errorf("RunScript: rm: No matches for glob: '%v'", trueFilepath)
					}

					for _, source := range trueFilepathGlob {
						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KRemoving files ... \t0%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Removing files\", \"details\": \"Removing %v\", \"progress\": 0, \"eta\": 0}", source)
						}

						var err error
						success := false
						for attempts := 0; attempts < 5; attempts++ {
							err = os.RemoveAll(source)
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
										// If the file(s) we are trying to delete is/are locked, retry 5 times with 250 millisecond intervals before finally crashing
										time.Sleep(250 * time.Millisecond)
										continue
									}
								} else {
									return fmt.Errorf("RunScript: rm: Unable to remove path: '%v': %v", source, err)
								}
							} else {
								return fmt.Errorf("RunScript: rm: Unable to remove path: '%v': %v", source, err)
							}
						}

						if !success {
							return fmt.Errorf("RunScript: rm: Unable to remove path: '%v': %v", source, err)
						}

						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KRemoving files ... \t100%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Removing files\", \"details\": \"Removing %v\", \"progress\": 100, \"eta\": 0}\n", source)
						}
					}
				}
			case "rsynca":
				// The rsynca command can take various pairs of arguments (path|path) separated by the '|' character and merges two folders overwriting the second one (destination)
				// It runs very similar to 'rsync -a', but we decided not to include rsync as a dependency for Windows users
				for i, rawFilepaths := range command {
					if i == 0 {
						continue
					}

					rawFilepathSource := strings.Split(rawFilepaths, "|")[0]
					rawFilepathDestination := strings.Split(rawFilepaths, "|")[1]

					filepathSource := GetProcessedFilePath(rawFilepathSource, gameMeta)
					filepathDestination := GetProcessedFilePath(rawFilepathDestination, gameMeta)

					filepathSourceGlob, err := filepath.Glob(filepathSource)
					if err != nil {
						return fmt.Errorf("RunScript: rsynca: Invalid glob syntax: '%v'", filepathSource)
					}

					if filepathSourceGlob == nil {
						return fmt.Errorf("RunScript: rsynca: No matches for glob: '%v'", filepathSource)
					}

					filepathDestinationGlob, err := filepath.Glob(filepathDestination)
					if err != nil {
						return fmt.Errorf("RunScript: rsynca: Invalid glob syntax: '%v'", filepathDestination)
					}

					if filepathDestinationGlob == nil {
						return fmt.Errorf("RunScript: rsynca: No matches for glob: '%v'", filepathDestination)
					}

					if len(filepathDestinationGlob) != 1 {
						return fmt.Errorf("RunScript: rsynca: More than one destination for glob: '%v'", filepathDestination)
					}

					sources := filepathSourceGlob
					destination := filepathDestinationGlob[0]

					for _, source := range sources {
						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KPatching game files ... \t0%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Patching game files\", \"details\": \"Patching %v on %v\", \"progress\": 0, \"eta\": 0}", source, destination)
						}

						err := RsyncA(source, destination)
						if err != nil {
							return err
						}

						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KPatching game files ... \t100%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Patching game files\", \"details\": \"Patching %v on %v\", \"progress\": 100, \"eta\": 0}\n", source, destination)
						}
					}
				}
			case "mv":
				// The mv command can take various pairs of arguments (path|path) separated by the '|' character and moves a file or a directory
				// Mostly identical to the Unix mv command
				for i, rawFilepaths := range command {
					if i == 0 {
						continue
					}

					rawFilepathSource := strings.Split(rawFilepaths, "|")[0]
					rawFilepathDestination := strings.Split(rawFilepaths, "|")[1]

					filepathSource := GetProcessedFilePath(rawFilepathSource, gameMeta)
					filepathDestination := GetProcessedFilePath(rawFilepathDestination, gameMeta)

					filepathSourceGlob, err := filepath.Glob(filepathSource)
					if err != nil {
						return fmt.Errorf("RunScript: mv: Invalid glob syntax: '%v'", filepathSource)
					}

					if filepathSourceGlob == nil {
						return fmt.Errorf("RunScript: mv: No matches for glob: '%v'", filepathSource)
					}

					filepathDestinationGlob, err := filepath.Glob(filepathDestination)
					if err != nil {
						return fmt.Errorf("RunScript: mv: Invalid glob syntax: '%v'", filepathDestination)
					}

					if filepathDestinationGlob == nil {
						return fmt.Errorf("RunScript: mv: No matches for glob: '%v'", filepathDestination)
					}

					if len(filepathDestinationGlob) != 1 {
						return fmt.Errorf("RunScript: mv: More than one destination for glob: '%v'", filepathDestination)
					}

					sources := filepathSourceGlob
					destination := filepathDestinationGlob[0]

					destinationInfo, err := os.Stat(destination)
					if os.IsNotExist(err) {
						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KMoving files ... \t0%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Moving files\", \"details\": \"Moving %v to %v\", \"progress\": 0, \"eta\": 0}", sources[0], destination)
						}

						err := os.Rename(sources[0], destination)
						if err != nil {
							return fmt.Errorf("RunScript: mv: Unable to rename paths: '%v' -> '%v': %v", sources[0], destination, err)
						}

						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KMoving files ... \t100%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Moving files\", \"details\": \"Moving %v to %v\", \"progress\": 100, \"eta\": 0}\n", sources[0], destination)
						}

						continue
					} else if err != nil {
						return fmt.Errorf("RunScript: mv: Unable to inspect path: '%v'", destination)
					}

					if !destinationInfo.IsDir() {
						return fmt.Errorf("RunScript: mv: Destination is not a directory")
					}

					for _, source := range sources {
						target := filepath.Join(destination, filepath.Base(source))

						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KMoving files ... \t0%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Moving files\", \"details\": \"Moving %v to %v\", \"progress\": 0, \"eta\": 0}", source, target)
						}

						err = os.Rename(source, target)
						if err != nil {
							return fmt.Errorf("RunScript: mv: Unable to rename paths: '%v' -> '%v': %v", source, target, err)
						}

						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KMoving files ... \t100%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Moving files\", \"details\": \"Moving %v to %v\", \"progress\": 100, \"eta\": 0}\n", source, target)
						}
					}
				}
			}
		}

		progress++
		
		err = WriteProgress(progress, gameMeta)
		if err != nil {
			return err
		}
	}

	return nil
}