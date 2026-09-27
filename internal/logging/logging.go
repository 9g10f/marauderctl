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

type Logger struct {
	File 	*os.File
	Location string
}

func (l Logger) Write(p []byte) (int, error) {
	_, err := Log(string(p), l.Location, "STD", l.File)
	return len(p), err
}

func (l Logger) Log(message string, level string) (int, error) {
	return Log(message, l.Location, level, l.File)
}

func (l Logger) LogError(err error, errorid string) error {
	return LogError(err, l.Location, errorid, l.File)
}

func Log(message string, location string, level string, logfile *os.File) (int, error) {
	if message != "" {
		var i int
		for txt := range strings.SplitSeq(message, "\n") {
			j, err := fmt.Fprintf(logfile, "[%v] (%v) %v: %v\n", time.Now().Format("2006/01/02 15:04:05.000"), location, level, txt)

			i += j

			if err != nil {
				return i, err
			}
		}

		return i, nil
	} else {
		return 0, nil
	}
}

func LogError(err error, location string, errorid string, logfile *os.File) error {
	Log(err.Error() + " (" + errorid + ")", location, "ERROR", logfile)
	return fmt.Errorf("%v (%v)", err.Error(), errorid)
}

func GetInstallLogFile(st time.Time, gameMeta cstructs.GameMeta) (*os.File, error) {
	gameFolder := gameMeta.GetDirectory()

	filePath := filepath.Join(gameFolder, ".marauder.logs", "install", st.Format("20060102150405"))

	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return nil, err
	}

	var file *os.File

	if _, err := os.Stat(filePath); err == nil {
		file, err = os.OpenFile(filePath, os.O_APPEND | os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
	} else if errors.Is(err, os.ErrNotExist) {
		file, err = os.Create(filePath)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	
	return file, nil
}

func GetRuntimeLogFile(st time.Time, gameMeta cstructs.GameMeta) (*os.File, error) {
	gameFolder := gameMeta.GetDirectory()

	filePath := filepath.Join(gameFolder, ".marauder.logs", "runtime", st.Format("20060102150405"))

	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return nil, err
	}

	var file *os.File

	if _, err := os.Stat(filePath); err == nil {
		file, err = os.OpenFile(filePath, os.O_APPEND | os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
	} else if errors.Is(err, os.ErrNotExist) {
		file, err = os.Create(filePath)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	
	return file, nil
}