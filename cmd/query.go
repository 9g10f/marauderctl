package cmd

import (
	"errors"
	"fmt"
	"strings"

	"codeberg.org/9g10f/marauderctl/internal/script"
	"github.com/spf13/cobra"
)

func Query() *cobra.Command {
	var server string

	cmd := &cobra.Command{
		Use: "query <game-id>@[game-version] <query-param:script|name|latest-version|dir|exe|max-size|size>",
		Short: "Get information about a game",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[1] {
			case "script":
				gameScript, err := script.GetVersionedScript(args[0], server)
				if err != nil {
					return err
				}

				fmt.Print(strings.Join(gameScript, "\n"))
			case "name":
				gameScript, err := script.GetScript(args[0], server)
				if err != nil {
					return err
				}

				gameName, err := script.ParseGameName(gameScript)
				if err != nil {
					return err
				}

				fmt.Print(gameName)
			case "latest-version":
				gameScript, err := script.GetScript(args[0], server)
				if err != nil {
					return err
				}

				fmt.Print(script.ParseGameLatestVersion(gameScript))
			case "dir":
				gameScript, err := script.GetScript(args[0], server)
				if err != nil {
					return err
				}

				dir, err := script.ParseGameDir(gameScript)
				if err != nil {
					return err
				}

				fmt.Print(dir)
			case "exe":
				gameScript, err := script.GetScript(args[0], server)
				if err != nil {
					return err
				}

				exe, err := script.ParseGameExe(gameScript)
				if err != nil {
					return err
				}

				fmt.Print(exe)
			case "max-size":
				gameScript, err := script.GetScript(args[0], server)
				if err != nil {
					return err
				}

				masSize, err := script.ParseGameMaxSize(gameScript)
				if err != nil {
					return err
				}

				fmt.Print(masSize)
			case "size":
				gameScript, err := script.GetScript(args[0], server)
				if err != nil {
					return err
				}

				size, err := script.ParseGameSize(gameScript)
				if err != nil {
					return err
				}

				fmt.Print(size)
			default:
				return errors.New("Invalid query parameter")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&server, "server", "s", DEFAULTSERVER(), "Server with install gameScripts")

	return cmd
}