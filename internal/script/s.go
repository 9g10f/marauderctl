package script

import (
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
)

func GetScript(game string, server string) ([]string, error) {
	var fullScript []string

	gameId := GetGameId(game)

	if strings.Count(game, "@") > 1 {
		return nil, errors.New("Invalid game field format")
	}

	if before, ok := strings.CutPrefix(server, "file://"); ok {
		data, err := os.ReadFile(before + gameId)
		if err != nil {
			return nil, err
		}

		fullScript = strings.Split(string(data), "\n")
	} else {
		r, err := http.Get(server + gameId)
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()

		data, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		fullScript = strings.Split(string(data), "\n")
	}

	fullScript = ReformatScript(fullScript)

	err := ValidateScript(fullScript, game)
	if err != nil {
		return nil, err
	}

	return fullScript, nil
}