package cmd

import (
	"os/exec"

	"codeberg.org/9g10f/marauderctl/internal/script"
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
			gameScript, err := script.GetScript(args[0], server)
			if err != nil {
				return err
			}

			exe, err := script.ParseGameExe(gameScript)
			if err != nil {
				return err
			}

			exePath := script.GetProcessedFilePath(exe, installPath)

			exeCmd := exec.Command(exePath)
			
			err = exeCmd.Run()
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