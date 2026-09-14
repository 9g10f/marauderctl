package main

var Version = "0.0.2"

func main() {
	Root.AddCommand(Install())
	Root.AddCommand(Query())

	err := Root.Execute()
	if err != nil {
		panic(err)
	}
}