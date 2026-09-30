package logging

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
)

// This file handles both the install and runtime logging

type Logger struct {
	File 	*os.File
	Location string
}

func (l Logger) Write(p []byte) (int, error) {
	message := string(p)

	if message != "" {
		var i int
		for txt := range strings.SplitSeq(message, "\n") {
			j, err := fmt.Fprintf(l.File, "[%v] (%v) STD: %v\n", time.Now().Format("2006/01/02 15:04:05.000"), l.Location, txt)

			i += j

			if err != nil {
				return len(p), fmt.Errorf("Logger: Write: Unable to write to log file: %v", err)
			}
		}

		return len(p), nil
	} else {
		return 0, nil
	}
}

func GetInstallLogFile(st time.Time, gameMeta cstructs.GameMeta) (*os.File, error) {
	filePath := filepath.Join(gameMeta.GetInstallLogsDirectory(), st.Format("20060102150405"))

	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return nil, fmt.Errorf("GetInstallLogFile: Unable to create install logs directory: '%v'", filepath.Dir(filePath))
	}

	var file *os.File

	// If log file already exists (shouldn't naturally happen) the contents of the next writes are appended
	// If the log file doesn't exist (the most common) it is created
	if _, err := os.Stat(filePath); err == nil {
		file, err = os.OpenFile(filePath, os.O_APPEND | os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("GetInstallLogFile: Unable to open file: '%v'", filePath)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		file, err = os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("GetInstallLogFile: Unable to create file: '%v'", filePath)
		}
	} else {
		return nil, fmt.Errorf("GetInstallLogFile: Unable to inspect file: '%v'", filePath)
	}
	
	return file, nil
}

func GetRuntimeLogFile(st time.Time, gameMeta cstructs.GameMeta) (*os.File, error) {
	filePath := filepath.Join(gameMeta.GetRuntimeLogsDirectory(), st.Format("20060102150405"))

	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return nil, fmt.Errorf("GetRuntimeLogFile: Unable to create runtime logs directory: '%v'", filepath.Dir(filePath))
	}

	var file *os.File

	// If log file already exists (shouldn't naturally happen) the contents of the next writes are appended
	// If the log file doesn't exist (the most common) it is created
	if _, err := os.Stat(filePath); err == nil {
		file, err = os.OpenFile(filePath, os.O_APPEND | os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("GetRuntimeLogFile: Unable to open file: '%v'", filePath)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		file, err = os.Create(filePath)
		if err != nil {
			return nil, fmt.Errorf("GetRuntimeLogFile: Unable to create file: '%v'", filePath)
		}
	} else {
		return nil, fmt.Errorf("GetRuntimeLogFile: Unable to inspect file: '%v'", filePath)
	}
	
	return file, nil
}