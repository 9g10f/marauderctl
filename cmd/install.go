package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"codeberg.org/9g10f/marauderctl/internal/disk"
	"codeberg.org/9g10f/marauderctl/internal/script"
	"github.com/spf13/cobra"
)

func VerifyAvailableSize(gameMeta cstructs.GameMeta) error {
	freeSpace, err := disk.GetFreeSpace(filepath.VolumeName(gameMeta.InstallPath))
	if err != nil {
		return err
	}

	fullScript, err := script.GetScript(gameMeta)
	if err != nil {
		return err
	}

	vars := script.ParseScriptVariables(fullScript, gameMeta)

	maxSize, ok := vars["max-size"]
	if !ok {
		return nil
	}

	maxSizeI, err := strconv.Atoi(maxSize)
	if err != nil {
		return err
	}

	if freeSpace < maxSizeI {
		return errors.New("No space left on device")
	}

	return nil
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
		SilenceUsage: true,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if strings.ToLower(outputStyle) != "default" && strings.ToLower(outputStyle) != "json" && strings.ToLower(outputStyle) != "silent" {
				return fmt.Errorf("Invalid output type: '%v'", outputStyle)
			}

			fullScript, err := script.GetScript(gameMeta)
			if err != nil {
				return err
			}

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

			os.MkdirAll(gameMeta.GetDirectory(), 0755)

			err = script.RunScript(gameScript, installFlags, gameMeta)
			if err != nil {
				return err
			}

			env, err := json.Marshal(scriptVars)
			if err != nil {
				return err
			}

			localScriptFile, err := os.Create(gameMeta.GetEnvFile())
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
	cmd.Flags().StringVarP(&outputStyle, "output-style", "o", "default", "Output style")

	return cmd
}