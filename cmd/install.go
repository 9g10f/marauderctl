package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"codeberg.org/9g10f/marauderctl/internal/script"
	"github.com/spf13/cobra"
)

func VerifyAvailableSize(script []string, installPath string) error {
	return nil // TODO: Implement disk scanning
}

func Install() *cobra.Command {
	var server string
	var installPath string
	var force bool // Ignores the device's remaining size for installation
	var resumeline int
	var outputStyle string // Output style: Default, JSON, Silent

	cmd := &cobra.Command{
		Use: "install <game-id>@[game-version]",
		Short: "Install a game",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.ToLower(outputStyle) != "default" && strings.ToLower(outputStyle) != "json" && strings.ToLower(outputStyle) != "silent" {
				return fmt.Errorf("Invalid output type: '%v'", outputStyle)
			}

			fullScript, err := script.GetScript(args[0], server)
			if err != nil {
				return err
			}

			gameScript, err := script.GetVersionedScriptFromFullScript(args[0], fullScript)
			if err != nil {
				return err
			}

			scriptVars := script.ParseScriptVariables(args[0], fullScript)

			if !force {
				err = VerifyAvailableSize(gameScript, installPath)
				if err != nil {
					return err
				}
			}

			gameId := script.GetGameId(args[0])

			os.MkdirAll(filepath.Join(installPath, gameId), 0755)

			err = script.RunScript(gameScript, force, installPath, resumeline, outputStyle, args[0])
			if err != nil {
				return err
			}

			env, err := json.Marshal(scriptVars)
			if err != nil {
				return err
			}

			localScriptFile, err := os.Create(filepath.Join(installPath, gameId, ".marauder.env"))
			if err != nil {
				return err
			}

			localScriptFile.Write(env)
			localScriptFile.Close()

			return nil
		},
	}

	cmd.Flags().StringVarP(&server, "server", "s", DEFAULTSERVER(), "Server with install scripts")
	cmd.Flags().StringVarP(&installPath, "install-path", "p", DEFAULTINSTALLPATH(), "Folder path for game installation")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Ignore warnings and proceed with installation regardless")
	cmd.Flags().IntVarP(&resumeline, "resume-line", "r", 0, "Resume installation from a specified install script line number")
	cmd.Flags().StringVarP(&outputStyle, "output", "o", "default", "Output style")

	return cmd
}