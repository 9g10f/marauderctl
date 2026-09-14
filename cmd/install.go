package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Install() *cobra.Command {
	var server string

	cmd := &cobra.Command{
		Use: "install <game-id>@[game-version]",
		Short: "Install a game",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			script, err := GetVersionedScript(args[0], server)
			if err != nil {
				return err
			}

			fmt.Println(script)

			for linen := range script {
				fmt.Printf("\r%v/%v", linen, len(script))
			}

			fmt.Printf("\r%v/%v", len(script), len(script))

			return nil
		},
	}

	// cmd.Flags().StringVarP(&server, "server", "s", "https://marauder.k.vu/s/", "Server with install scripts")
	cmd.Flags().StringVarP(&server, "server", "s", "file://./s/", "Server with install scripts")

	return cmd
}