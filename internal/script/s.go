package script

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
)

func Search(text string, server string) ([]string, error) {
	searchURL := fmt.Sprintf("search?txt=%v")

	r, err := http.Get(path.Join(server, searchURL))
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

	return strings.Split(string(data), "\n"), nil
}

func GetScript(gameMeta cstructs.GameMeta) ([]string, error) {
	if strings.Count(gameMeta.Game, "@") > 1 {
		return nil, errors.New("Invalid game field format")
	}

	r, err := http.Get(path.Join(gameMeta.Server, gameMeta.GetId()))
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

	fullScript := strings.Split(string(data), "\n")

	fullScript = ReformatScript(fullScript)

	err = ValidateScript(fullScript)
	if err != nil {
		return nil, err
	}

	return fullScript, nil
}