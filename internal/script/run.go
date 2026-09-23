package script

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/buildkite/shellwords"
)

// ERROR_SHARING_VIOLATION is not exported by the standard syscall package on Windows
// Its value (32) is fixed by the Windows API
const ERROR_SHARING_VIOLATION syscall.Errno = 32

func RunScript(script []string, force bool, installPath string, resumeline int, outputStyle string) error {
	for linen, line := range script {
		if linen + 1 < resumeline {
			continue
		}

		command, _ := shellwords.Split(line)
		
		if len(command) > 0 { // Skip blank lines
			switch command[0] {
			case "set":
				// Variables are handled by ParseScriptVariables
				continue
			case "download":
				// The download command can take various arguments (URLs), of clearnet, BitTorrent and in the future Tor files
				// It downloads all files onto 'installPath' with the name on the URL
				for i, downloadURL := range command {
					if i == 0 {
						continue
					}

					if outputStyle == "default" {
						fmt.Printf("\r\033[2KDownloading game files ...   0%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00")
					} else if outputStyle == "json" {
						fmt.Printf("\r\033[2K{\"task\": \"Downloading game files\", \"details\": \"Downloading %v\", \"progress\": 0, \"eta\": 0}", downloadURL)
					}

					if strings.HasPrefix(downloadURL, "http:") || strings.HasPrefix(downloadURL, "https:") {
						err := DownloadHTTP(downloadURL, installPath)
						if err != nil {
							return err
						}
					} else {
						err := DownloadTorrentMagnet(downloadURL, installPath, outputStyle)
						if err != nil {
							return err
						}
					}

					if outputStyle == "default" {
						fmt.Printf("\r\033[2KDownloading game files ...   100%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
					} else if outputStyle == "json" {
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

					trueFilepath := GetProcessedFilePath(rawFilepath, installPath)

					trueFilepathGlob, _ := filepath.Glob(trueFilepath)
					if trueFilepathGlob == nil {
						return fmt.Errorf("File(s) not found: '%v'", trueFilepath)
					}

					for _, source := range trueFilepathGlob {
						if outputStyle == "default" {
							fmt.Printf("\r\033[Unzipping game files ...   0%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00")
						} else if outputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Unzipping game files\", \"details\": \"Unzipping %v\", \"progress\": 0, \"eta\": 0}", source)
						}

						command := exec.Command(
							"7z",
							"x",
							source,
							"-ponline-fix.me", // TODO: Needs to be customizable in the future
							"-y",
						)

						command.Dir = filepath.Dir(source)

						err := command.Run()
						if err != nil {
							return err
						}

						if outputStyle == "default" {
							fmt.Printf("\r\033[Unzipping game files ...   100%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
						} else if outputStyle == "json" {
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

					trueFilepath := GetProcessedFilePath(rawFilepath, installPath)

					trueFilepathGlob, _ := filepath.Glob(trueFilepath)
					if trueFilepathGlob == nil {
						return fmt.Errorf("File(s) not found: '%v'", trueFilepath)
					}

					for _, source := range trueFilepathGlob {
						if outputStyle == "default" {
							fmt.Printf("\r\033[Removing files ...   0%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00")
						} else if outputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Removing files\", \"details\": \"Removing %v\", \"progress\": 0, \"eta\": 0}", source)
						}

						var err error
						success := false
						for attempts := 0; attempts < 40; attempts++ {
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

						if outputStyle == "default" {
							fmt.Printf("\r\033[Removing files ...   100%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
						} else if outputStyle == "json" {
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

					for _, source := range sources {
						if outputStyle == "default" {
							fmt.Printf("\r\033[Patching game files ...   0%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
						} else if outputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Patching game files\", \"details\": \"Patching %v on %v\", \"progress\": 0, \"eta\": 0}\n", source, destination)
						}

						err := RsyncA(source, destination)
						if err != nil {
							return err
						}

						if outputStyle == "default" {
							fmt.Printf("\r\033[Patching game files ...   100%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
						} else if outputStyle == "json" {
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
						if outputStyle == "default" {
							fmt.Printf("\r\033[Moving files ...   0%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
						} else if outputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Moving files\", \"details\": \"Moving %v to %v\", \"progress\": 0, \"eta\": 0}\n", sources[0], destination)
						}

						if err := os.Rename(sources[0], destination); err != nil {
							return err
						}

						if outputStyle == "default" {
							fmt.Printf("\r\033[Moving files ...   100%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
						} else if outputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Moving files\", \"details\": \"Moving %v to %v\", \"progress\": 100, \"eta\": 0}\n", sources[0], destination)
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

						if outputStyle == "default" {
							fmt.Printf("\r\033[Moving files ...   0%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
						} else if outputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Moving files\", \"details\": \"Moving %v to %v\", \"progress\": 0, \"eta\": 0}\n", source, target)
						}

						err = os.Rename(source, target)
						if err != nil {
							return err
						}

						if outputStyle == "default" {
							fmt.Printf("\r\033[Moving files ...   100%% (0 / 0 bytes) @ 0 MiB/s ETA 0:00:00\n")
						} else if outputStyle == "json" {
							fmt.Printf("\r\033[2K{\"task\": \"Moving files\", \"details\": \"Moving %v to %v\", \"progress\": 100, \"eta\": 0}\n", source, target)
						}
					}
				}
			}
		}
	}

	return nil
}