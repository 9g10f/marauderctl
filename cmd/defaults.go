package cmd

import (
	"os"
	"path/filepath"
)

func DEFAULTINSTALLPATH() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	return filepath.Join(homeDir, "Marauder Games")
}

func DEFAULTSERVER() string {
	return "https://marauder.k.vu/"
}