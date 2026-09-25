package cmd

import (
	"os"

	"codeberg.org/9g10f/marauderctl/internal/script"
	"github.com/spf13/cobra"
)

func Uninstall() *cobra.Command {
	var server string
	var installPath string
	
	cmd := &cobra.Command{
		Use: "uninstall <game-id>@[game-version]",
		Short: "Uninstall a game",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return os.RemoveAll(script.GetGameFolder(installPath, server, args[0]))
		},
	}

	cmd.Flags().StringVarP(&server, "server", "s", DEFAULTSERVER(), "Server with install scripts")
	cmd.Flags().StringVarP(&installPath, "install-path", "p", DEFAULTINSTALLPATH(), "Folder path for game installation")

	return cmd
}