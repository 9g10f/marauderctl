package cmd

import (
	"fmt"
	"strings"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"codeberg.org/9g10f/marauderctl/internal/script"
	"github.com/spf13/cobra"
)

// The query command lets you fetch information or environment variables about a specified game

func Query() *cobra.Command {
	var server string
	var installPath string

	cmd := &cobra.Command{
		Use: "query <game-id>@[game-version] <query-param:script|latest-version|*>",
		Short: "Get information about a game",
		Args: cobra.ExactArgs(2),
		SilenceUsage: true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			gameMeta := cstructs.GameMeta{
				Game: args[0],
				InstallPath: installPath,
				Server: server,
			}
			
			switch args[1] {
			case "script":
				fullScript, err := script.GetScript(gameMeta)
				if err != nil {
					return err
				}

				gameScript, err := script.GetVersionedScript(fullScript, gameMeta)
				if err != nil {
					return err
				}

				fmt.Println(strings.Join(gameScript, "\n"))
			case "latest-version":
				gameScript, err := script.GetScript(gameMeta)
				if err != nil {
					return err
				}

				fmt.Println(script.ParseGameLatestVersion(gameScript))
			default:
				vars, err := script.ParseLocalScriptVariables(gameMeta)
				if err != nil {
					return err
				}

				fmt.Println(vars[args[1]])
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&installPath, "install-path", "p", DEFAULTINSTALLPATH(), "Folder path with game installation")
	cmd.Flags().StringVarP(&server, "server", "s", DEFAULTSERVER(), "Server with install gameScripts")

	return cmd
}