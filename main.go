package main

import (
	"log"

	"github.com/justinpopa/goreleaser-tfpp/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil {
		log.Panicln(err)
	}
}
