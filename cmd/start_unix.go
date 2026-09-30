//go:build unix

package cmd

import (
	"errors"
	"fmt"
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

// The start commands lets you start an installed game from it's game-id and game-version
// The Unix version supports Wine and Proton launching

func Start() *cobra.Command {
	var server string
	var installPath string
	var compatibilityLayer string
	var protonPath string

	cmd := &cobra.Command{
		Use: "start <game-id>@[game-version]",
		Short: "Start a game",
		SilenceUsage: true,
		SilenceErrors: true,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if 'compatibility-layer' is valid
			if compatibilityLayer != "none" && compatibilityLayer != "wine" && compatibilityLayer != "proton" {
				return fmt.Errorf("Start: Unsupported compatibility layer: '%v'", compatibilityLayer)
			}

			gameMeta := cstructs.GameMeta{
				Game: args[0],
				InstallPath: installPath,
				Server: server,
			}

			st := time.Now()

			// marauderctl saves runtime logs
			logfile, err := logging.GetRuntimeLogFile(st, gameMeta)
			if err != nil {
				return err
			}
			defer logfile.Close()

			logger := logging.Logger{
				File: logfile,
				Location: "game",
			}

			vars, err := script.ParseLocalScriptVariables(gameMeta)
			if err != nil {
				return err
			}

			// Even tho all enviroment variables are optional for install you can't start games that have no 'exe' variable
			exe, ok := vars["exe"]
			if !ok {
				return errors.New("Start: Unable to start the game, no EXE was found in the script variables")
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
					// Find installed Proton versions and run the best one
					proton, err = start.GetProton()
					if err != nil {
						return err
					}
				} else {
					proton = protonPath
				}

				command, err = start.ProtonRun(exePath, proton, gameMeta)
				if err != nil {
					return err
				}
			}

			// Redirect STDOUT & STDERR to the log
			command.Stdout = logger
			command.Stderr = logger

			err = command.Start()
			if err != nil {
				return errors.New("Start: Unable to start game: " + err.Error())
			}

			// While the game is running marauderctl redirects interrupt signals to the game

			signals := make(chan os.Signal, 1)
			signal.Notify(signals, os.Interrupt)

			var userInterrupted atomic.Bool

			go func() {
				<-signals
				userInterrupted.Store(true)
				syscall.Kill(-command.Process.Pid, syscall.SIGINT)
				fmt.Println("Game killed")
			}()

			err = command.Wait()
			if err != nil && !userInterrupted.Load() {
				return errors.New("Start: Game crashed: " + err.Error())
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