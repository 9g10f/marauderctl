package main

var Version = "0.0.1"

func main() {
	Root.AddCommand(Install)

	err := Root.Execute()
	if err != nil {
		panic(err)
	}
}