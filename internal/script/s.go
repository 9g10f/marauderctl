package script

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
)

func Search(text string, server string) ([]string, error) {
	searchURL := "search?txt=" + text

	r, err := http.Get(server + searchURL)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		return nil, fmt.Errorf("Error while searching games: response code was %v", r.StatusCode)
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var sResponse cstructs.SResponseSearch
	err = json.Unmarshal(data, &sResponse)
	if err != nil {
		return nil, err
	}

	if sResponse.Error != nil {
		return nil, errors.New(*sResponse.Error)
	}

	return sResponse.Results, nil
}

func GetScript(gameMeta cstructs.GameMeta) ([]string, error) {
	if strings.Count(gameMeta.Game, "@") > 1 {
		return nil, errors.New("Invalid game field format")
	}

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

	var sResponse cstructs.SResponseS

	err = json.Unmarshal(data, &sResponse)
	if err != nil {
		return nil, err
	}

	if sResponse.Error != nil {
		return nil, errors.New(*sResponse.Error)
	}

	fullScript := strings.Split(sResponse.Script, "\n")

	fullScript = ReformatScript(fullScript)

	err = ValidateScript(fullScript)
	if err != nil {
		return nil, err
	}

	return fullScript, nil
}