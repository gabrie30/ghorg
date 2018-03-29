package main

import (
	"fmt"
	"log"

	"github.com/ghorg/cmd"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Hello, Ghorg")
	cmd.CloneAllReposByOrg()
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
