package main

var Version = "0.0.2"

func marauderctl() error {
	Root.AddCommand(Install())
	Root.AddCommand(Query())

	err := Root.Execute()
	if err != nil {
		return err
	}

	return nil
}