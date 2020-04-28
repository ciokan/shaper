package main

import (
	"github.com/ciokan/shaper/cmd"
)

var (
	version = "dev"
)

func main() {
	cmd.Execute(version)
}
