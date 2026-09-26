package main

import (
	"os"

	"codeberg.org/9g10f/marauderctl/cmd"
)

func main() {
	err := cmd.MarauderCtl()
	if err != nil {
		os.Exit(1)
	}
}