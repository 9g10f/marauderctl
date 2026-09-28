package cmd

import (
	"encoding/json"
	"fmt"

	"codeberg.org/9g10f/marauderctl/internal/cstructs"
	"codeberg.org/9g10f/marauderctl/internal/script"
	"github.com/spf13/cobra"
)

func Search() *cobra.Command {
	var server string
	var outputStyle string // Default, JSON, Silent

	cmd := &cobra.Command{
		Use: "search <*>",
		Short: "Search for a game",
		SilenceUsage: true,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			results, err := script.Search(args[0], server)
			if err != nil {
				return err
			}

			if outputStyle == "default" {
				for _, result := range results {
					fullScript, err := script.GetScript(cstructs.GameMeta{Game: result, Server: server})
					if err != nil {
						return err
					}

					vars := script.ParseGlobalScriptVariables(fullScript)

					name := vars["name"]
					if name != "" {
						name = " | " + name
					}

					description := vars["description"]
					if description != "" {
						description = " | " + description
					}

					fmt.Println(result + name + description)
				}
			} else if outputStyle == "json" {
				raw, err := json.Marshal(results)
				if err != nil {
					return err
				}

				fmt.Println(string(raw))
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&server, "server", "s", DEFAULTSERVER(), "Server with install scripts")
	cmd.Flags().StringVarP(&outputStyle, "output-style", "o", "default", "Output style")

	return cmd
}