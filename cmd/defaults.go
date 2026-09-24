package cmd

import (
	"os"
)

func DEFAULTINSTALLPATH() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	return homeDir + "/Marauder/Games"
}

func DEFAULTSERVER() string {
	return "https://marauder.k.vu/s/"
}