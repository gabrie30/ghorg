package main

import (
	"fmt"
	"os"

	"github.com/ghorg/cmd"
)

func main() {
	fmt.Println("Hello, Ghorg")
	cmd.Clone(os.Args[1])
}
