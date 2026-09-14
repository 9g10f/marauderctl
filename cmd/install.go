package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Install = &cobra.Command{
	Use: 	"install <game>",
	Short: 	"Install a specified game",
	Args: 	cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Installing %v\n", args[0])
	},
}