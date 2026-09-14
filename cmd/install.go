package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func Install() *cobra.Command {
	var server string
	var force bool // Ignores the device's remaining size for installation

	cmd := &cobra.Command{
		Use: "install <game-id>@[game-version]",
		Short: "Install a game",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			script, err := GetVersionedScript(args[0], server)
			if err != nil {
				return err
			}

			fmt.Println(strings.Join(script, "\n"))

			err = RunScript(script, force)
			if err != nil {
				return err
			}

			return nil
		},
	}

	// cmd.Flags().StringVarP(&server, "server", "s", "https://marauder.k.vu/s/", "Server with install scripts")
	cmd.Flags().StringVarP(&server, "server", "s", "file://./s/", "Server with install scripts")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Ignore warnings and proceed with installation regardless")

	return cmd
}