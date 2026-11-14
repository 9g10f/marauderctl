package cmd

import (
	"errors"
	"path/filepath"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// The library command lets you create / edit Marauder libraries (aka install paths)

func Library() *cobra.Command {
	var installPath string

	cmd := &cobra.Command{
		Use: "library <command>",
		Short: "Modifies a Marauder library (install path)",
		SilenceUsage: true,
		SilenceErrors: true,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "create":
				err := os.Mkdir(installPath, 0755)
				if err != nil {
					if !errors.Is(err, os.ErrExist) {
						return fmt.Errorf("Library: Unable to create provided path: %v", err)
					}
				}

				if _, err := os.Stat(filepath.Join(installPath, ".marauder.ipath")); err == nil {
					return errors.New("Library: Provided path is already a Marauder library")
				} else if errors.Is(err, os.ErrNotExist) {
					ipathFile, err := os.Create(filepath.Join(installPath, ".marauder.ipath"))
					if err != nil {
						return fmt.Errorf("Library: Unable to create ipath file: '%v'", filepath.Join(installPath, ".marauder.ipath"))
					}

					err = ipathFile.Close()
					if err != nil {
						return fmt.Errorf("Library: Unable to close ipath file: '%v'", filepath.Join(installPath, ".marauder.ipath"))
					}
				} else {
					return fmt.Errorf("Library: Unable to analyse ipath file: '%v': %v", filepath.Join(installPath, ".marauder.ipath"), err)
				}
			case "disable":
				if _, err := os.Stat(filepath.Join(installPath, ".marauder.ipath")); err == nil {
					err := os.Remove(filepath.Join(installPath, ".marauder.ipath"))
					if err != nil {
						return fmt.Errorf("Library: Unable to remove ipath file: %v", err)
					}
				} else if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("Library: Provided path isn't a Marauder library")
				} else {
					return fmt.Errorf("Library: Unable to analyse ipath file: '%v': %v", filepath.Join(installPath, ".marauder.ipath"), err)
				}
			case "remove":
				if _, err := os.Stat(filepath.Join(installPath, ".marauder.ipath")); err == nil {
					err := os.Remove(filepath.Join(installPath, ".marauder.ipath"))
					if err != nil {
						return fmt.Errorf("Library: Unable to remove ipath file: %v", err)
					}
				} else if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("Library: Provided path isn't a Marauder library")
				} else {
					return fmt.Errorf("Library: Unable to analyse ipath file: '%v': %v", filepath.Join(installPath, ".marauder.ipath"), err)
				}

				err := os.RemoveAll(installPath)
				if err != nil {
					return fmt.Errorf("Library: Unable to remove game files: %v", err)
				}
			case "clean":
				// Remove all game files but leave the library
			}
			
			return nil
		},
	}

	cmd.Flags().StringVarP(&installPath, "install-path", "p", DEFAULTINSTALLPATH(), "Folder path for game installation")

	return cmd
}
