package main

import (
	"github.com/ciokan/shaper/cmd"
	"log"
)

var (
	version = "dev"
)

func main() {
	log.Fatal(cmd.Execute(version))
}
