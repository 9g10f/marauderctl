//go:build windows

package cmd

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync/atomic"
	"time"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"codeberg.org/9g10f/marauderctl/internal/logging"
	"codeberg.org/9g10f/marauderctl/internal/script"
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
			gameMeta := cstructs.GameMeta{
				Game: args[0],
				InstallPath: installPath,
				Server: server,
			}

			st := time.Now()

			logfile, err := logging.GetRuntimeLogFile(st, gameMeta)
			if err != nil {
				return err
			}

			logger := logging.Logger{
				File: logfile,
				Location: "game",
			}

			mainLogger := logging.Logger{
				File: logfile,
				Location: "main",
			}
			
			vars, err := script.ParseLocalScriptVariables(gameMeta)
			if err != nil {
				return mainLogger.LogError(err, "0")
			}

			exe, ok := vars["exe"]
			if !ok {
				return mainLogger.LogError(errors.New("Unable to start the game, no EXE was found in the script variables"), "1")
			}

			exePath := script.GetProcessedFilePath(exe, gameMeta)

			command := exec.Command(exePath)
			command.Dir = filepath.Dir(exePath)
			command.Stdout = logger
			command.Stderr = logger

			err = command.Start()
			if err != nil {
				return mainLogger.LogError(err, "2")
			}

			signals := make(chan os.Signal, 1)
			signal.Notify(signals, os.Interrupt)

			var userInterrupted atomic.Bool

			go func() {
				<-signals
				userInterrupted.Store(true)
				command.Process.Kill()
			}()

			err = command.Wait()
			if err != nil && !userInterrupted.Load() {
				return mainLogger.LogError(err, "3")
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