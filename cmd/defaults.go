package cmd

import (
	"os"
	"path/filepath"
)

// This file contains the defaults for marauderctl
// If you don't pass any custom flags these are the used values

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