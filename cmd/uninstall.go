package cmd

import (
	"fmt"
	"os"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"github.com/spf13/cobra"
)

// The uninstall command lets you uninstall games from a Marauder library

func Uninstall() *cobra.Command {
	var server string
	var installPath string
	
	cmd := &cobra.Command{
		Use: "uninstall <game-id>@[game-version]",
		Short: "Uninstall a game",
		SilenceUsage: true,
		SilenceErrors: true,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			gameMeta := cstructs.GameMeta{
				Game: args[0],
				InstallPath: installPath,
				Server: server,
			}

			err := os.RemoveAll(gameMeta.GetDirectory())
			if err != nil {
				return fmt.Errorf("Uninstall: Unable to remove game install directory: %v", err)
			}
			
			return nil
		},
	}

	cmd.Flags().StringVarP(&server, "server", "s", DEFAULTSERVER(), "Server with install scripts")
	cmd.Flags().StringVarP(&installPath, "install-path", "p", DEFAULTINSTALLPATH(), "Folder path with game installation")

	return cmd
}