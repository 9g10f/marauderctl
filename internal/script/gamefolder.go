package script

import (
	"encoding/base64"
	"path/filepath"
	"strings"
)

func GetGameFolder(installPath string, server string, game string) string {
	gameId := GetGameId(game)
	gameVersion := GetGameVersion(game)

	return filepath.Join(installPath, base64.URLEncoding.EncodeToString([]byte(server)), gameId, gameVersion)
}

func GetProcessedFilePath(path string, installPath string, server string, game string) string {
	return strings.ReplaceAll(path, "$path", GetGameFolder(installPath, server, game))
}