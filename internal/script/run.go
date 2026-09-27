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

func RunScript(script []string, installFlags cstructs.InstallFlags, gameMeta cstructs.GameMeta) error {
	st := time.Now()

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

	mainLogger := logging.Logger{
		File: logfile,
		Location: "main",
	}

	_, ok := os.LookupEnv("TORRENT_STORAGE_DEFAULT_FILE_IO")
	if !ok {
		os.Setenv("TORRENT_STORAGE_DEFAULT_FILE_IO", "classic") // Set torrent setting to close mapped files
	}

	progress, err := GetProgress(gameMeta)
	if err != nil {
		progress = 0
	} else {
		installFlags.ResumeLine = progress + 1
	}
	
	WriteProgress(progress, gameMeta)

	for linen, line := range script {
		if linen + 1 < installFlags.ResumeLine {
			continue
		}

		command, _ := shellwords.Split(line)
		
		if len(command) > 0 { // Skip blank lines
			switch command[0] {
			case "set":
				// Variables are handled by ParseScriptVariables
				progress++
				WriteProgress(progress, gameMeta)

				continue
			case "download":
				// The download command can take various arguments (URLs), of clearnet, BitTorrent and in the future Tor files
				// It downloads all files onto 'installPath' with the name on the URL
				for i, downloadURL := range command {
					if i == 0 {
						continue
					}

					if installFlags.OutputStyle == "default" {
						fmt.Printf("\r\033[2KDownloading game files ...   0%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00")
					} else if installFlags.OutputStyle == "json" {
						fmt.Printf("\r\033[2K{\"task\": \"Downloading game files\", \"details\": \"Downloading %v\", \"progress\": 0, \"eta\": 0}", downloadURL)
					}

					if strings.HasPrefix(downloadURL, "magnet:") {
						mainLogger.Log(fmt.Sprintf("Downloading game files from '%v' with BitTorrent", downloadURL), "INFO")

						err := DownloadTorrentMagnet(downloadURL, installFlags.OutputStyle, torrentLogger, gameMeta)
						if err != nil {
							return mainLogger.LogError(err, "0")
						}

						mainLogger.Log(fmt.Sprintf("Finished downloading game files from '%v' with BitTorrent", downloadURL), "INFO")
					} else {
						mainLogger.Log(fmt.Sprintf("Downloading game files from '%v'", downloadURL), "INFO")

						err := Download(downloadURL, gameMeta)
						if err != nil {
							return mainLogger.LogError(err, "1")
						}

						mainLogger.Log(fmt.Sprintf("Finished downloading game files from '%v'", downloadURL), "INFO")
					}

					if installFlags.OutputStyle == "default" {
						fmt.Printf("\r\033[2KDownloading game files ...   100%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
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

					trueFilepathGlob, _ := filepath.Glob(trueFilepath)
					if trueFilepathGlob == nil {
						return mainLogger.LogError(fmt.Errorf("File(s) not found: '%v'", trueFilepath), "2")
					}

					for _, source := range trueFilepathGlob {
						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KUnzipping game files ...   0%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Unzipping game files\", \"details\": \"Unzipping %v\", \"progress\": 0, \"eta\": 0}", source)
						}

						mainLogger.Log(fmt.Sprintf("Unzipping '%v' using 7z", source), "INFO")

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
								return mainLogger.LogError(err, "3")
							}

							code := exitError.ExitCode()

							// We don't throw an error when the exit code is 2 because that's the exit code 7z throws when even tho it still unzipped the files there were some warnings (e.g. Unsupported Method)
							if code != 2 {
								return mainLogger.LogError(err, "4")
							}
						}

						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KUnzipping game files ...   100%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Unzipping game files\", \"details\": \"Unzipping %v\", \"progress\": 100, \"eta\": 0}\n", source)
						}

						mainLogger.Log(fmt.Sprintf("Finished unzipping '%v' using 7z", source), "INFO")
					}
				}
			case "rm":
				// The rm command can take various arguments (paths) and removes all files/directories
				for i, rawFilepath := range command {
					if i == 0 {
						continue
					}

					trueFilepath := GetProcessedFilePath(rawFilepath, gameMeta)

					trueFilepathGlob, _ := filepath.Glob(trueFilepath)
					if trueFilepathGlob == nil {
						return mainLogger.LogError(fmt.Errorf("File(s) not found: '%v'", trueFilepath), "5")
					}

					for _, source := range trueFilepathGlob {
						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KRemoving files ...   0%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Removing files\", \"details\": \"Removing %v\", \"progress\": 0, \"eta\": 0}", source)
						}

						mainLogger.Log(fmt.Sprintf("Removing '%v'", source), "INFO")

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
								}
							}

							return mainLogger.LogError(err, "6")
						}

						if !success {
							return mainLogger.LogError(err, "7")
						}

						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KRemoving files ...   100%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Removing files\", \"details\": \"Removing %v\", \"progress\": 100, \"eta\": 0}\n", source)
						}

						mainLogger.Log(fmt.Sprintf("Finished removing '%v'", source), "INFO")
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

					filepathSourceGlob, _ := filepath.Glob(filepathSource)
					if filepathSourceGlob == nil {
						return mainLogger.LogError(fmt.Errorf("File(s) not found: '%v'", filepathSource), "8")
					}

					filepathDestinationGlob, _ := filepath.Glob(filepathDestination)
					if filepathDestinationGlob == nil {
						return mainLogger.LogError(fmt.Errorf("File(s) not found: '%v'", filepathDestination), "9")
					}

					if len(filepathDestinationGlob) != 1 {
						return mainLogger.LogError(fmt.Errorf("Too many destinations: '%v'", filepathDestination), "10")
					}

					sources := filepathSourceGlob
					destination := filepathDestinationGlob[0]

					for _, source := range sources {
						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KPatching game files ...   0%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Patching game files\", \"details\": \"Patching %v on %v\", \"progress\": 0, \"eta\": 0}", source, destination)
						}

						mainLogger.Log(fmt.Sprintf("Rsyncing '%v' on '%v'", source, destination), "INFO")

						err := RsyncA(source, destination)
						if err != nil {
							return mainLogger.LogError(err, "11")
						}

						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KPatching game files ...   100%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Patching game files\", \"details\": \"Patching %v on %v\", \"progress\": 100, \"eta\": 0}\n", source, destination)
						}

						mainLogger.Log(fmt.Sprintf("Finshed rsyncing '%v' on '%v'", source, destination), "INFO")
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

					filepathSourceGlob, _ := filepath.Glob(filepathSource)
					if filepathSourceGlob == nil {
						return mainLogger.LogError(fmt.Errorf("File(s) not found: '%v'", filepathSource), "12")
					}

					filepathDestinationGlob, _ := filepath.Glob(filepathDestination)
					if filepathDestinationGlob == nil {
						return mainLogger.LogError(fmt.Errorf("File(s) not found: '%v'", filepathDestination), "13")
					}

					if len(filepathDestinationGlob) != 1 {
						return mainLogger.LogError(fmt.Errorf("Too many destinations: '%v'", filepathDestination), "14")
					}

					sources := filepathSourceGlob
					destination := filepathDestinationGlob[0]

					destinationInfo, err := os.Stat(destination)
					if os.IsNotExist(err) {
						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KMoving files ...   0%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Moving files\", \"details\": \"Moving %v to %v\", \"progress\": 0, \"eta\": 0}", sources[0], destination)
						}

						mainLogger.Log(fmt.Sprintf("Moving '%v' to '%v'", sources[0], destination), "INFO")

						if err := os.Rename(sources[0], destination); err != nil {
							return mainLogger.LogError(err, "15")
						}

						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KMoving files ...   100%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Moving files\", \"details\": \"Moving %v to %v\", \"progress\": 100, \"eta\": 0}\n", sources[0], destination)
						}

						mainLogger.Log(fmt.Sprintf("Finished moving '%v' to '%v'", sources[0], destination), "INFO")

						continue
					} else if err != nil {
						return mainLogger.LogError(err, "16")
					}

					if !destinationInfo.IsDir() {
						return mainLogger.LogError(fmt.Errorf("Destination is not a directory: '%v'", destination), "17")
					}

					for _, source := range sources {
						target := filepath.Join(destination, filepath.Base(source))

						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KMoving files ...   0%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Moving files\", \"details\": \"Moving %v to %v\", \"progress\": 0, \"eta\": 0}", source, target)
						}

						mainLogger.Log(fmt.Sprintf("Moving '%v' to '%v'", source, target), "INFO")

						err = os.Rename(source, target)
						if err != nil {
							return mainLogger.LogError(err, "18")
						}

						mainLogger.Log(fmt.Sprintf("Finished moving '%v' to '%v'", source, target), "INFO")

						if installFlags.OutputStyle == "default" {
							fmt.Printf("\r\033[2KMoving files ...   100%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
						} else if installFlags.OutputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Moving files\", \"details\": \"Moving %v to %v\", \"progress\": 100, \"eta\": 0}\n", source, target)
						}
					}
				}
			}
		}

		progress++
		WriteProgress(progress, gameMeta)
	}

	return nil
}