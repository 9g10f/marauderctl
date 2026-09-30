package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"codeberg.org/9g10f/marauderctl/internal/disk"
	"codeberg.org/9g10f/marauderctl/internal/script"
	"github.com/spf13/cobra"
)

// The update command lets you update the latest version of a game to the most recent provided by the server
// Because of the nature of pirated games and it's shipping methods the most efficient way of updating is most times simply uninstalling the current version and installing the new one

func Update() *cobra.Command {
	var server string
	var installPath string
	var force bool // Ignores the device's remaining size for installation
	var resumeline int
	var outputStyle string // Default, JSON, Silent

	cmd := &cobra.Command{
		Use: "update <game-id>",
		Short: "Update an installed game to the latest version",
		SilenceUsage: true,
		SilenceErrors: true,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if 'output-style' is valid
			if strings.ToLower(outputStyle) != "default" && strings.ToLower(outputStyle) != "json" && strings.ToLower(outputStyle) != "silent" {
				return fmt.Errorf("Update: Invalid output type: '%v'", outputStyle)
			}

			gameMeta := cstructs.GameMeta{
				Game: args[0],
				InstallPath: installPath,
				Server: server,
			}

			installFlags := cstructs.InstallFlags{
				Force: force,
				ResumeLine: resumeline,
				OutputStyle: outputStyle,
			}

			// If the game wasn't installed with @latest it can't be updated
			if gameMeta.GetVersion() != "latest" {
				return errors.New("Update: You can only update the latest version of a game")
			}

			fullScript, err := script.GetScript(gameMeta)
			if err != nil {
				return err
			}

			latestVersion := script.ParseGameLatestVersion(fullScript)

			// If there already exists a version file we check if it's up to date
			// If there isn't a version file then either the game was never installed (in which case we install it) or it was installing and crashed (which also means to install it by resuming)
			versionFilePath := gameMeta.GetVersionFile()
			if _, err := os.Stat(versionFilePath); err == nil {
				versionFile, err := os.Open(versionFilePath)
				if err != nil {
					return fmt.Errorf("Update: Unable to open file: '%v'", versionFilePath)
				}

				versionBytes, err := io.ReadAll(versionFile)
				if err != nil {
					return fmt.Errorf("Update: Unable to read file: '%v'", versionFilePath)
				}

				version := string(versionBytes)

				if version == latestVersion {
					return errors.New("Update: Game is up to date")
				}
			} else if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("Update: Unable to inspect version file: '%v'", versionFilePath)
			}

			// Uninstall the game
			err = os.RemoveAll(gameMeta.GetDirectory())
			if err != nil {
				return fmt.Errorf("Update: Unable to remove game install directory: %v", err)
			}

			// Install it again
			gameScript, err := script.GetVersionedScript(fullScript, gameMeta)
			if err != nil {
				return err
			}

			scriptVars := script.ParseScriptVariables(fullScript, gameMeta)

			if !force {
				err = disk.VerifyAvailableSize(gameMeta)
				if err != nil {
					return err
				}
			}

			err = os.MkdirAll(gameMeta.GetDirectory(), 0755)
			if err != nil {
				return fmt.Errorf("Update: Unable to create game install directory: '%v'", gameMeta.GetDirectory())
			}

			ipathFile, err := os.Create(filepath.Join(installPath, ".marauder.ipath"))
			if err != nil {
				return fmt.Errorf("Update: Unable to create ipath file: '%v'", filepath.Join(installPath, ".marauder.ipath"))
			}
			
			err = ipathFile.Close()
			if err != nil {
				return fmt.Errorf("Update: Unable to close ipath file: '%v'", filepath.Join(installPath, ".marauder.ipath"))
			}

			err = script.RunScript(gameScript, installFlags, gameMeta)
			if err != nil {
				return err
			}

			env, err := json.Marshal(scriptVars)
			if err != nil {
				return errors.New("Update: Unable to parse environment variables")
			}

			envFile, err := os.Create(gameMeta.GetEnvFile())
			if err != nil {
				return fmt.Errorf("Update: Unable to create environment file: '%v'", gameMeta.GetEnvFile())
			}
			defer envFile.Close()

			_, err = envFile.Write(env)
			if err != nil {
				return fmt.Errorf("Update: Unable to write to environment file: '%v'", gameMeta.GetEnvFile())
			}

			versionFile, err := os.Create(gameMeta.GetVersionFile())
			if err != nil {
				return fmt.Errorf("Update: Unable to create version file: '%v'", gameMeta.GetVersionFile())
			}
			defer versionFile.Close()

			_, err = versionFile.Write([]byte(script.ParseGameLatestVersion(fullScript)))
			if err != nil {
				return fmt.Errorf("Update: Unable to write to version file: '%v'", gameMeta.GetVersionFile())
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&server, "server", "s", DEFAULTSERVER(), "Server with install scripts")
	cmd.Flags().StringVarP(&installPath, "install-path", "p", DEFAULTINSTALLPATH(), "Folder path for game installation")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Ignore warnings and proceed with installation regardless")
	cmd.Flags().IntVarP(&resumeline, "resume-line", "r", 0, "Resume installation from a specified install script line number")
	cmd.Flags().StringVarP(&outputStyle, "output-style", "o", "default", "Output style")

	return cmd
}