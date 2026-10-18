package script

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"codeberg.org/9g10f/marauderctl/internal/logging"
)

var progressRe = regexp.MustCompile(`^\s*(\d{1,3})%`)

func GetProgressFrom7zLine(line []byte) (int, bool) {
    m := progressRe.FindSubmatch(line)
    if m == nil {
        return 0, false
    }

    p, err := strconv.Atoi(string(m[1]))
    if err != nil {
        return 0, false
    }

    return p, true
}

func HandleLine(line []byte, start time.Time, source string, installFlags cstructs.InstallFlags) {
	completed, ok := GetProgressFrom7zLine(line)
	if !ok {
		return
	}

	elapsed := time.Since(start).Seconds()
	var avgSpeed float64

	if elapsed > 0 {
		avgSpeed = float64(completed) / elapsed
	}

	var eta int
	if avgSpeed > 0 {
		eta = int(float64(100 - completed) / avgSpeed)
	}

	hours := eta / 3600
	minutes := (eta % 3600) / 60
	seconds := eta % 60

	switch installFlags.OutputStyle {
	case "default":
		fmt.Printf("\r\033[2KUnzipping game files ... \t%d%% @ %.2f %%/s ETA %d:%02d:%02d", completed, avgSpeed, hours, minutes, seconds)
	case "json":
		fmt.Printf("\r\033[2K{\"task\": \"Unzipping game files\", \"details\": \"Unzipping %v\", \"progress\": %d, \"eta\": %d}", source, completed, eta)
	}
}

func Unzip(source string, unzipLogger logging.Logger, installFlags cstructs.InstallFlags) error {
	command := exec.Command(
		"7z",
		"x",
		source,
		"-ponline-fix.me",
		"-y",
		"-bsp1",
	)

	command.Dir = filepath.Dir(source)
	command.Stderr = unzipLogger

	stdout, err := command.StdoutPipe()
	if err != nil {
		return fmt.Errorf("Unzip: Error while starting 7z: %v", err)
	}

	err = command.Start()
	if err != nil {
		return fmt.Errorf("Unzip: Error while starting 7z: %v", err)
	}

	reader := bufio.NewReader(stdout)
	var line []byte
	start := time.Now()

	for {
		b, err := reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				command.Process.Kill()
				return fmt.Errorf("Unzip: Error while running 7z: %v", err)
			}
		}

		_, err = unzipLogger.Write([]byte{b})
		if err != nil {
			command.Process.Kill()
			return fmt.Errorf("Unzip: Error while running 7z: %v", err)
		}

		switch b {
		case '\r', '\n', '\b':
			HandleLine(line, start, source, installFlags)
			line = []byte{}
		default:
			line = append(line, b)
		}
	}

	err = command.Wait()
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if !ok {
			return fmt.Errorf("Unzip: Error while running 7z: %v", err)
		}

		code := exitError.ExitCode()

		// We don't throw an error when the exit code is 2 because that's the exit code 7z throws when even tho it still unzipped the files there were some warnings (e.g. Unsupported Method)
		if code != 2 {
			return fmt.Errorf("Unzip: Error while running 7z: %v", err)
		}
	}

	return nil
}