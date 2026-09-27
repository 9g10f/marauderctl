package cstructs

import (
	"encoding/base64"
	"path/filepath"
	"strings"
)

type GameMeta struct {
	Game 		string
	InstallPath string
	Server		string
}

func (g GameMeta) GetVersion() string {
	if !strings.Contains(g.Game, "@") {
		return "latest"
	} else {
		return strings.Split(g.Game, "@")[1]
	}
}

func (g GameMeta) GetId() string {
	if !strings.Contains(g.Game, "@") {
		return g.Game
	} else {
		return strings.Split(g.Game, "@")[0]
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

type InstallFlags struct {
	Force 		bool
	ResumeLine	int
	OutputStyle	string
}