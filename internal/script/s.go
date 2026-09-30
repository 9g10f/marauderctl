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

// Returns the search results as a lits of game-ids
func Search(text string, server string) ([]string, error) {
	searchURL := "search?txt=" + text

	r, err := http.Get(server + searchURL)
	if err != nil {
		return nil, fmt.Errorf("Search: Unable to make request to server: %v", err)
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		return nil, fmt.Errorf("Search: Error while searching games: response code was %v", r.StatusCode)
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("Search: Unable to read response's body: %v", err)
	}

	var sResponse cstructs.SResponseSearch
	err = json.Unmarshal(data, &sResponse)
	if err != nil {
		return nil, fmt.Errorf("Search: Unable to parse response: %v", err)
	}

	if sResponse.Error != nil {
		return nil, fmt.Errorf("Search: Server returned an error: %v", *sResponse.Error)
	}

	return sResponse.Results, nil
}

// Returns the game script
func GetScript(gameMeta cstructs.GameMeta) ([]string, error) {
	if strings.Count(gameMeta.Game, "@") > 1 {
		return nil, errors.New("GetScript: Invalid game field format")
	}

	r, err := http.Get(gameMeta.Server + gameMeta.GetId())
	if err != nil {
		return nil, fmt.Errorf("GetScript: Unable to make request to server: %v", err)
	}
	defer r.Body.Close()

	if r.StatusCode != 200 {
		return nil, fmt.Errorf("GetScript: Error while downloading install script: response code was %v", r.StatusCode)
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("GetScript: Unable to read response's body: %v", err)
	}

	var sResponse cstructs.SResponseS
	err = json.Unmarshal(data, &sResponse)
	if err != nil {
		return nil, fmt.Errorf("GetScript: Unable to parse response: %v", err)
	}

	if sResponse.Error != nil {
		return nil, fmt.Errorf("GetScript: Server returned an error: %v", *sResponse.Error)
	}

	fullScript := strings.Split(sResponse.Script, "\n")
	fullScript = ReformatScript(fullScript)

	err = ValidateScript(fullScript)
	if err != nil {
		return nil, err
	}

	return fullScript, nil
}