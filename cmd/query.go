package cmd

import (
	"errors"
	"fmt"
	"strings"

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
				script, err := GetVersionedScript(args[0], server)
				if err != nil {
					return err
				}

				fmt.Print(strings.Join(script, "\n"))
			case "name":
				script, err := GetScript(args[0], server)
				if err != nil {
					return err
				}

				gameName, err := ParseGameName(script)
				if err != nil {
					return err
				}

				fmt.Print(gameName)
			case "latest-version":
				script, err := GetScript(args[0], server)
				if err != nil {
					return err
				}

				fmt.Print(ParseGameLatestVersion(script))
			case "dir":
				script, err := GetScript(args[0], server)
				if err != nil {
					return err
				}

				dir, err := ParseGameDir(script)
				if err != nil {
					return err
				}

				fmt.Print(dir)
			case "exe":
				script, err := GetScript(args[0], server)
				if err != nil {
					return err
				}

				exe, err := ParseGameExe(script)
				if err != nil {
					return err
				}

				fmt.Print(exe)
			case "max-size":
				script, err := GetScript(args[0], server)
				if err != nil {
					return err
				}

				masSize, err := ParseGameMaxSize(script)
				if err != nil {
					return err
				}

				fmt.Print(masSize)
			case "size":
				script, err := GetScript(args[0], server)
				if err != nil {
					return err
				}

				size, err := ParseGameSize(script)
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

	cmd.Flags().StringVarP(&server, "server", "s", DEFAULTSERVER(), "Server with install scripts")

	return cmd
}