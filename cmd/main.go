package main

func main() {
	err := marauderctl()
	if err != nil {
		panic(err)
	}
}