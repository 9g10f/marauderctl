package main

import (
	"fmt"
	"os"

	"codeberg.org/9g10f/marauderctl/cmd"
)

// This is marauderctl's main entry point

func main() {
	err := cmd.MarauderCtl()
	if err != nil {
		// Every exit that's not a clean exit returns an error code 1 and an error message
		
		fmt.Fprintf(os.Stderr, "marauderctl: %v\n", err)
		os.Exit(1)
	}
}