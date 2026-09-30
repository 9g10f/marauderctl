package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// The list command lets you check which games and game versions are currently installed in a specified library path

func List() *cobra.Command {
	var installPath string

	cmd := &cobra.Command{
		Use: "list",
		Short: "List all installed games",
		SilenceUsage: true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var l []string

			// Check for the ipath file
			if _, err := os.Stat(filepath.Join(installPath, ".marauder.ipath")); err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("List: Provided path is not a Marauder install path: '%v'", installPath)
				} else {
					return fmt.Errorf("List: Unable to inspect ipath file: '%v'", filepath.Join(installPath, ".marauder.ipath"))
				}
			}

			if _, err := os.Stat(installPath); err == nil {
				servers, err := os.ReadDir(installPath)
				if err != nil {
					return fmt.Errorf("List: Unable to list install path directory: '%v'", installPath)
				}

				for _, server := range servers {
					if !server.IsDir() {
						continue
					}

					games, err := os.ReadDir(filepath.Join(installPath, server.Name()))
					if err != nil {
						return fmt.Errorf("List: Unable to list server directory: '%v'", filepath.Join(installPath, server.Name()))
					}

					for _, game := range games {
						if !server.IsDir() {
							continue
						}

						versions, err := os.ReadDir(filepath.Join(installPath, server.Name(), game.Name()))
						if err != nil {
							return fmt.Errorf("List: Unable to list game directory: '%v'", filepath.Join(installPath, server.Name(), game.Name()))
						}

						for _, version := range versions {
							if !server.IsDir() {
								continue
							}

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
				return fmt.Errorf("List: Unable to inspect install path directory: '%v'", installPath)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&installPath, "install-path", "p", DEFAULTINSTALLPATH(), "Folder path for game installation")

	return cmd
}