package cmd

import (
	"os/exec"

	"github.com/spf13/cobra"
)

func Start() *cobra.Command {
	var server string
	var installPath string

	cmd := &cobra.Command{
		Use: "start <game-id>@[game-version]",
		Short: "Start a game",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			script, err := GetScript(args[0], server)
			if err != nil {
				return err
			}

			exe, err := ParseGameExe(script)
			if err != nil {
				return err
			}

			exePath := GetProcessedFilePath(exe, installPath)

			exeCmd := exec.Command(exePath)
			
			err = exeCmd.Start()
			if err != nil {
				return err
			}

			return nil
		},
	}
	
	cmd.Flags().StringVarP(&server, "server", "s", DEFAULTSERVER(), "Server with install scripts")
	cmd.Flags().StringVarP(&installPath, "install-path", "p", DEFAULTINSTALLPATH(), "Folder path for game installation")

	return cmd
}