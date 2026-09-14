package main

import (
	"io"
	"net/http"
	"os"
	"strings"
)

func GetInstallScript(game string, server string) ([]string, error) {
	var script []string

	if before, ok := strings.CutSuffix(server, "file://"); ok {
		data, err := os.ReadFile(before + game)
		if err != nil {
			return nil, err
		}

		script = strings.Split(string(data), "\n")
	} else {
		r, err := http.Get(server + game)
		if err != nil {
			return nil, err
		}
		defer r.Body.Close()

		data, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}

		script = strings.Split(string(data), "\n")
	}

	return script, nil
}