package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Install() *cobra.Command {
	var server string

	cmd := &cobra.Command{
		Use: 	"install <game>",
		Short: 	"Install a specified game",
		Args: 	cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			script, err := GetInstallScript(args[0], server)
			if err != nil {
				panic(err)
			}

			for linen := range script {
				fmt.Printf("\r%v/%v", linen, len(script))
			}
		},
	}

	// cmd.Flags().StringVarP(&server, "server", "s", "https://marauder.k.vu/s/", "Server with install scripts")
	cmd.Flags().StringVarP(&server, "server", "s", "file://./s/", "Server with install scripts")

	return cmd
}