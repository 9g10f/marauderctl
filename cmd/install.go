package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"codeberg.org/9g10f/marauderctl/internal/disk"
	"codeberg.org/9g10f/marauderctl/internal/script"
	"github.com/spf13/cobra"
)

// The install command installs games with the install script fetched from the server with it's game-id and game-version on a specified library path
// It also checks for available space on install

func Install() *cobra.Command {
	var server string
	var installPath string
	var force bool // Ignores the device's remaining size for installation
	var resumeline int
	var outputStyle string // Default, JSON, Silent

	cmd := &cobra.Command{
		Use: "install <game-id>@[game-version]",
		Short: "Install a game",
		SilenceUsage: true,
		SilenceErrors: true,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if 'output-style' is valid
			if strings.ToLower(outputStyle) != "default" && strings.ToLower(outputStyle) != "json" && strings.ToLower(outputStyle) != "silent" {
				return fmt.Errorf("Install: Invalid output type: '%v'", outputStyle)
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

			fullScript, err := script.GetScript(gameMeta)
			if err != nil {
				return err
			}

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
				return fmt.Errorf("Install: Unable to create game install directory: '%v'", gameMeta.GetDirectory())
			}

			// The ipath file is used to find/verify Marauder install paths
			ipathFile, err := os.Create(filepath.Join(installPath, ".marauder.ipath"))
			if err != nil {
				return fmt.Errorf("Install: Unable to create ipath file: '%v'", filepath.Join(installPath, ".marauder.ipath"))
			}

			err = ipathFile.Close()
			if err != nil {
				return fmt.Errorf("Install: Unable to close ipath file: '%v'", filepath.Join(installPath, ".marauder.ipath"))
			}

			err = script.RunScript(gameScript, installFlags, gameMeta)
			if err != nil {
				return err
			}

			env, err := json.Marshal(scriptVars)
			if err != nil {
				return errors.New("Install: Unable to parse environment variables")
			}

			envFile, err := os.Create(gameMeta.GetEnvFile())
			if err != nil {
				return fmt.Errorf("Install: Unable to create environment file: '%v'", gameMeta.GetEnvFile())
			}
			defer envFile.Close()

			_, err = envFile.Write(env)
			if err != nil {
				return fmt.Errorf("Install: Unable to write to environment file: '%v'", gameMeta.GetEnvFile())
			}

			if gameMeta.GetVersion() == "latest" {
				// Installing a game with @latest and @<version> isn't the same even if <version> is the latest
				// Installing a game with @latest makes it updatable, you can't run 'marauderctl update' on a game installed with a specified version
				// Because of that marauderctl needs to keep track of what version is installed in latest/, that's what the version file is for
				versionFile, err := os.Create(gameMeta.GetVersionFile())
				if err != nil {
					return fmt.Errorf("Install: Unable to create version file: '%v'", gameMeta.GetVersionFile())
				}
				defer versionFile.Close()

				_, err = versionFile.Write([]byte(script.ParseGameLatestVersion(fullScript)))
				if err != nil {
					return fmt.Errorf("Install: Unable to write to version file: '%v'", gameMeta.GetVersionFile())
				}
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