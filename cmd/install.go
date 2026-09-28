package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"codeberg.org/9g10f/marauderctl/internal/disk"
	"codeberg.org/9g10f/marauderctl/internal/script"
	"github.com/spf13/cobra"
)

func root(p string) string {
	p = filepath.Clean(p)
	for {
		if _, err := os.Stat(p); err == nil {
			return p
		}
		parent := filepath.Dir(p)
		if parent == p {
			return p
		}
		p = parent
	}
}

func VerifyAvailableSize(gameMeta cstructs.GameMeta) error {
	freeSpace, err := disk.GetFreeSpace(root(gameMeta.InstallPath))
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

	// Give +10 MB of room
	if freeSpace < (maxSizeI + int(math.Pow(10, 7))) {
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
			defer envFile.Close()

			_, err = envFile.Write(env)
			if err != nil {
				return err
			}

			if gameMeta.GetVersion() == "latest" {
				versionFile, err := os.Create(gameMeta.GetVersionFile())
				if err != nil {
					return err
				}
				defer versionFile.Close()

				_, err = versionFile.Write([]byte(script.ParseGameLatestVersion(fullScript)))
				if err != nil {
					return err
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