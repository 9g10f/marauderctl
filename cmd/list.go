package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func List() *cobra.Command {
	var installPath string

	cmd := &cobra.Command{
		Use: "list",
		Short: "List all installed games",
		RunE: func(cmd *cobra.Command, args []string) error {
			var l []string

			if _, err := os.Stat(installPath); err == nil {
				servers, err := os.ReadDir(installPath)
				if err != nil {
					return err
				}

				for _, server := range servers {
					games, err := os.ReadDir(filepath.Join(installPath, server.Name()))
					if err != nil {
						return err
					}

					for _, game := range games {
						versions, err := os.ReadDir(filepath.Join(installPath, server.Name(), game.Name()))
						if err != nil {
							return err
						}

						for _, version := range versions {
							l = append(l, game.Name() + "@" + version.Name())
						}
					}
				}

				for _, g := range l {
					fmt.Println(g)
				}
			} else if errors.Is(err, os.ErrNotExist) {
				return nil
			} else {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&installPath, "install-path", "p", DEFAULTINSTALLPATH(), "Folder path for game installation")

	return cmd
}