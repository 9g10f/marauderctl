package cmd

var Version = "0.0.15"

func MarauderCtl() error {
	Root.AddCommand(Install())
	Root.AddCommand(Query())
	Root.AddCommand(Start())
	Root.AddCommand(Uninstall())
	Root.AddCommand(List())
	Root.AddCommand(Update())

	err := Root.Execute()
	if err != nil {
		return err
	}

	return nil
}