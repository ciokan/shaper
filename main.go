package main

import (
	"log"

	"github.com/ciokan/shaper/cmd"
)

var (
	version = "dev"
)

func main() {
	log.Fatal(cmd.Execute(version))
}
