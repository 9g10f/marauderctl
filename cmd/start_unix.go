//go:build unix

package cmd

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"codeberg.org/9g10f/marauderctl/internal/logging"
	"codeberg.org/9g10f/marauderctl/internal/script"
	"codeberg.org/9g10f/marauderctl/internal/start"
	"github.com/spf13/cobra"
)

func Start() *cobra.Command {
	var server string
	var installPath string
	var compatibilityLayer string
	var protonPath string

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

			gameLogger := logging.Logger{
				File: logfile,
				Location: "game",
			}

			mainLogger := logging.Logger{
				File: logfile,
				Location: "main",
			}
			
			if compatibilityLayer != "none" && compatibilityLayer != "wine" && compatibilityLayer != "proton" {
				return mainLogger.LogError(errors.New("Unsupported compatibility layer"), "0")
			}

			vars, err := script.ParseLocalScriptVariables(gameMeta)
			if err != nil {
				return mainLogger.LogError(err, "1")
			}

			exe, ok := vars["exe"]
			if !ok {
				return mainLogger.LogError(errors.New("Unable to start the game, no EXE was found in the script variables"), "2")
			}

			exePath := script.GetProcessedFilePath(exe, gameMeta)

			var command *exec.Cmd

			switch compatibilityLayer {
			case "none":
				command = start.RawRun(exePath)
			case "wine":
				command = start.WineRun(exePath)
			case "proton":
				var proton string
				if protonPath == "" {
					proton, err = start.GetProton()
					if err != nil {
						return mainLogger.LogError(err, "3")
					}
				} else {
					proton = protonPath
				}

				command, err = start.ProtonRun(exePath, proton, gameMeta)
				if err != nil {
					return mainLogger.LogError(err, "4")
				}
			}

			command.Stdout = gameLogger
			command.Stderr = gameLogger

			err = command.Start()
			if err != nil {
				return mainLogger.LogError(err, "5")
			}

			signals := make(chan os.Signal, 1)
			signal.Notify(signals, os.Interrupt)

			var userInterrupted atomic.Bool

			go func() {
				<-signals
				userInterrupted.Store(true)
				syscall.Kill(-command.Process.Pid, syscall.SIGINT)
			}()

			err = command.Wait()
			if err != nil && !userInterrupted.Load() {
				return mainLogger.LogError(err, "6")
			}

			signal.Stop(signals)
			close(signals)

			return nil
		},
	}
	
	cmd.Flags().StringVarP(&server, "server", "s", DEFAULTSERVER(), "Server with install scripts")
	cmd.Flags().StringVarP(&installPath, "install-path", "p", DEFAULTINSTALLPATH(), "Folder path with game installation")
	cmd.Flags().StringVarP(&compatibilityLayer, "compatibility", "c", "proton", "Compatibility layer")
	cmd.Flags().StringVar(&protonPath, "proton-path", "", "Custom Proton path")

	return cmd
}