package script

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
)

func GetScript(gameMeta cstructs.GameMeta) ([]string, error) {
	var fullScript []string

	if strings.Count(gameMeta.Game, "@") > 1 {
		return nil, errors.New("Invalid game field format")
	}

	if before, ok := strings.CutPrefix(gameMeta.Server, "file://"); ok {
		data, err := os.ReadFile(before + gameMeta.GetId())
		if err != nil {
			return nil, err
		}

		fullScript = strings.Split(string(data), "\n")
	} else {
		r, err := http.Get(gameMeta.Server + gameMeta.GetId())
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()

		if r.StatusCode != 200 {
			return nil, fmt.Errorf("Error while downloading install script: response code was %v", r.StatusCode)
		}

		data, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		fullScript = strings.Split(string(data), "\n")
	}

	fullScript = ReformatScript(fullScript)

	err := ValidateScript(fullScript)
	if err != nil {
		return nil, err
	}

	return fullScript, nil
}