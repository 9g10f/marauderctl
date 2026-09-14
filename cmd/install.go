package main

import (
	"github.com/spf13/cobra"
)

func VerifyAvailableSize(script []string, installPath string) error {
	return nil
}

func Install() *cobra.Command {
	var server string
	var installPath string
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

			if !force {
				err = VerifyAvailableSize(script, installPath)
				if err != nil {
					return err
				}
			}

			err = RunScript(script, force, installPath)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&server, "server", "s", DEFAULTSERVER(), "Server with install scripts")
	cmd.Flags().StringVarP(&installPath, "install-path", "p", DEFAULTINSTALLPATH(), "Ignore warnings and proceed with installation regardless")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Ignore warnings and proceed with installation regardless")

	return cmd
}