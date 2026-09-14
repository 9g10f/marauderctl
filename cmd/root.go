package main

import (
	"github.com/spf13/cobra"
)

var Root = &cobra.Command{
	Use: "marauderctl",
	Version: Version,
	Short: "The CLI backend for Marauder Launcher",
}