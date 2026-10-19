package cstructs

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// The GameMeta struct is used as an arugment to most script-related functions, it holds the game-id, game-version, install path and server
// It also provides quality-of-life functions such as GetId, GetVersion, GetServerBase64 (used for the folder names), GetDirectory (game's install path), GetEnvFile, GetVersionFile, GetProgressFile, GetInstallLogsDirectory and GetRuntimeLogsDirectory
type GameMeta struct {
	Game 		string
	InstallPath string
	Server		string
}

func (g GameMeta) GetId() string {
	if !strings.Contains(g.Game, "@") {
		return g.Game
	} else {
		return strings.Split(g.Game, "@")[0]
	}
}

func (g GameMeta) GetVersion() string {
	if !strings.Contains(g.Game, "@") {
		return "latest"
	} else {
		return strings.Split(g.Game, "@")[1]
	}
}

func (g GameMeta) GetServerBase64() string {
	return base64.URLEncoding.EncodeToString([]byte(g.Server))
}

func (g GameMeta) GetDirectory() string {
	return filepath.Join(g.InstallPath, g.GetServerBase64(), g.GetId(), g.GetVersion())
}

func (g GameMeta) GetEnvFile() string {
	return filepath.Join(g.GetDirectory(), ".marauder.env")
}

func (g GameMeta) GetVersionFile() string {
	return filepath.Join(g.GetDirectory(), ".marauder.version")
}

func (g GameMeta) GetProgressFile() string {
	return filepath.Join(g.GetDirectory(), ".marauder.iprogress")
}

func (g GameMeta) GetInstallLogsDirectory() string {
	return filepath.Join(g.GetDirectory(), ".marauder.logs", "install")
}

func (g GameMeta) GetRuntimeLogsDirectory() string {
	return filepath.Join(g.GetDirectory(), ".marauder.logs", "runtime")
}

func (g GameMeta) GetDoneFile() string {
	return filepath.Join(g.GetDirectory(), ".marauder.idone")
}

func (g GameMeta) SetDone() error {
	file, err := os.Create(g.GetDoneFile())
	if err != nil {
		return fmt.Errorf("GameMeta: SetDone: Unable to create done file: %v", err)
	}

	err = file.Close()
	if err != nil {
		return fmt.Errorf("GameMeta: SetDone: Unable to close done file: %v", err)
	}

	return nil
}

func (g GameMeta) SetNotDone() error {
	err := os.Remove(g.GetDoneFile())
	if err != nil {
		return fmt.Errorf("GameMeta: SetNotDone: Unable to remove done file: %v", err)
	}
	
	return nil
}

func (g GameMeta) IsDone() (bool, error) {
	if _, err := os.Stat(g.GetDoneFile()); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, fmt.Errorf("GameMeta: IsDone: Unable to inspect file: '%v'", g.GetDoneFile())
	}
}

type InstallFlags struct {
	Force 		bool
	ResumeLine	int
	OutputStyle	string
	NoSeeding	bool
}