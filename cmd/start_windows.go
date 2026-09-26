//go:build windows

package cmd

import (
	"errors"
	"os"
	"os/signal"

	"codeberg.org/9g10f/marauderctl/internal/script"
	"codeberg.org/9g10f/marauderctl/internal/start"
	"github.com/spf13/cobra"
)

func Start() *cobra.Command {
	var server string
	var installPath string

	cmd := &cobra.Command{
		Use: "start <game-id>@[game-version]",
		Short: "Start a game",
		SilenceUsage: true,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vars, err := script.ParseLocalScriptVariables(args[0], installPath, server)
			if err != nil {
				return err
			}

			exe, ok := vars["exe"]
			if !ok {
				return errors.New("Unable to start the game, no EXE was found in the script variables")
			}

			exePath := script.GetProcessedFilePath(exe, installPath, server, args[0])

			exeCmd := start.RawRun(exePath)

			err = exeCmd.Start()
			if err != nil {
				return err
			}

			signals := make(chan os.Signal, 1)
			signal.Notify(signals, os.Interrupt)

			go func() {
				<-signals
				exeCmd.Process.Kill()
			}()

			err = exeCmd.Wait()
			if err != nil {
				return err
			}

			signal.Stop(signals)
			close(signals)

			return nil
		},
	}
	
	cmd.Flags().StringVarP(&server, "server", "s", DEFAULTSERVER(), "Server with install scripts")
	cmd.Flags().StringVarP(&installPath, "install-path", "p", DEFAULTINSTALLPATH(), "Folder path with game installation")

	return cmd
}