package main

import (
	"log"
	"os"
	"os/exec"
	"testing"
)

func TestMain(m *testing.M) {
	err := exec.Command("go", "run", "main.go", "version").Run()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}
