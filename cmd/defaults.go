package cmd

import (
	"os"
	"runtime"
)

func DEFAULTINSTALLPATH() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERPROFILE") + "/Marauder/Games"
	} else {
		return os.Getenv("HOME") + "/Marauder/Games"
	}
}

func DEFAULTSERVER() string {
	return "file://./s/"
	// https://marauder.k.vu/s/
}