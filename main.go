package main

import (
	"fmt"
	"log"

	"github.com/gabrie30/ghorg/cmd"
	"github.com/joho/godotenv"
	homedir "github.com/mitchellh/go-homedir"
)

func main() {
	fmt.Println("Hello, Ghorg")
	cmd.CloneAllReposByOrg()
}

func init() {
	home, err := homedir.Dir()

	if err != nil {
		log.Fatal("Error trying to find users home directory")
	}

	err = godotenv.Load(home + "/.ghorg")
	if err != nil {
		log.Fatal("Error loading .ghorg file, create a .env from the sample and run Make install")
	}
}
