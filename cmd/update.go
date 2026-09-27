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
	"codeberg.org/9g10f/marauderctl/internal/script"
	"github.com/spf13/cobra"
)

func Update() *cobra.Command {
	var server string
	var installPath string
	var force bool // Ignores the device's remaining size for installation
	var resumeline int
	var outputStyle string // Output style: Default, JSON, Silent

	cmd := &cobra.Command{
		Use: "update <game-id>",
		Short: "Update an installed game to the latest version",
		SilenceUsage: true,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.ToLower(outputStyle) != "default" && strings.ToLower(outputStyle) != "json" && strings.ToLower(outputStyle) != "silent" {
				return fmt.Errorf("Invalid output type: '%v'", outputStyle)
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

			if gameMeta.GetVersion() != "latest" {
				return errors.New("You can only update the latest version of a game")
			}

			fullScript, err := script.GetScript(gameMeta)
			if err != nil {
				return err
			}

			latestVersion := script.ParseGameLatestVersion(fullScript)

			// If there already exists a version file we check if it's up to date
			// If there isn't a version file then either the game was never installed (in which case we install it) or it was installing and crashed (which also means to install it by resuming)
			if _, err := os.Stat(gameMeta.GetVersionFile()); err == nil {
				versionFile, err := os.Open(gameMeta.GetVersionFile())
				if err != nil {
					return err
				}

				versionBytes, err := io.ReadAll(versionFile)
				if err != nil {
					return err
				}

				version := string(versionBytes)

				if version == latestVersion {
					return errors.New("Game is up to date")
				}
			} else if !errors.Is(err, os.ErrNotExist) {
				return err
			}

			os.RemoveAll(gameMeta.GetDirectory())

			gameScript, err := script.GetVersionedScriptFromFullScript(fullScript, gameMeta)
			if err != nil {
				return err
			}

			scriptVars := script.ParseScriptVariables(fullScript, gameMeta)

			if !force {
				err = VerifyAvailableSize(gameMeta)
				if err != nil {
					return err
				}
			}

			err = os.MkdirAll(gameMeta.GetDirectory(), 0755)
			if err != nil {
				return err
			}

			ipathFile, err := os.Create(filepath.Join(installPath, ".marauder.ipath"))
			if err != nil {
				return err
			}
			ipathFile.Close()

			err = script.RunScript(gameScript, installFlags, gameMeta)
			if err != nil {
				return err
			}

			env, err := json.Marshal(scriptVars)
			if err != nil {
				return err
			}

			envFile, err := os.Create(gameMeta.GetEnvFile())
			if err != nil {
				return err
			}

			envFile.Write(env)
			envFile.Close()

			versionFile, err := os.Create(gameMeta.GetVersionFile())
			if err != nil {
				return err
			}
			defer versionFile.Close()

			_, err = versionFile.Write([]byte(script.ParseGameLatestVersion(fullScript)))
			if err != nil {
				return err
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