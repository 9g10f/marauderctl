package cmd

var Version = "0.0.13"

func MarauderCtl() error {
	Root.AddCommand(Install())
	Root.AddCommand(Query())
	Root.AddCommand(Start())
	Root.AddCommand(Uninstall())
	Root.AddCommand(List())

	err := Root.Execute()
	if err != nil {
		return err
	}

	return nil
}