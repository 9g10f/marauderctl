package main

import "github.com/spf13/cobra"

func Query() *cobra.Command {
	cmd := &cobra.Command{
		Use: "query <game-id>@[game-version] <query-param:name|latest-version|dir|exe|max-size|size>",
		Short: "Get information about a game",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}