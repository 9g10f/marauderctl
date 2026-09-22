package cmd

var Version = "0.0.2"

func MarauderCtl() error {
	Root.AddCommand(Install())
	Root.AddCommand(Query())
	Root.AddCommand(Start())

	err := Root.Execute()
	if err != nil {
		return err
	}

	return nil
}