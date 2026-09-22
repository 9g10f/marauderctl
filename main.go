package main

import (
	"codeberg.org/9g10f/marauderctl/cmd"
)

func main() {
	err := cmd.MarauderCtl()
	if err != nil {
		panic(err)
	}
}