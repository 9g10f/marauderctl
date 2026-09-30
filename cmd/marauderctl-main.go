package cmd

// This file defines MarauderCtl, controls the version and organizes all commands

var Version = "0.0.17"

func MarauderCtl() error {
	Root.AddCommand(Install())
	Root.AddCommand(Query())
	Root.AddCommand(Start())
	Root.AddCommand(Uninstall())
	Root.AddCommand(List())
	Root.AddCommand(Update())
	Root.AddCommand(Search())

	return Root.Execute()
}