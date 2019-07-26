package main

import (
	"github.com/gabrie30/ghorg/cmd"
	"github.com/gabrie30/ghorg/configs"
)

func main() {
	configs.Load()
	cmd.Execute()
}
