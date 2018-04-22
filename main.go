package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gabrie30/ghorg/cmd"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Hello, Ghorg")
	cmd.CloneAllReposByOrg()
}

func init() {
	home := os.Getenv("HOME")
	err := godotenv.Load(home + "/.ghorg")
	if err != nil {
		log.Fatal("Error loading .ghorg file, create a .env from the sample and run Make install")
	}
}
