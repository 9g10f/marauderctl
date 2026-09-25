package cmd

var Version = "0.0.10"

func MarauderCtl() error {
	Root.AddCommand(Install())
	Root.AddCommand(Query())
	Root.AddCommand(Start())
	Root.AddCommand(Uninstall())

	err := Root.Execute()
	if err != nil {
		return err
	}

	return nil
}