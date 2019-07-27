package main

import (
	"os"
	"testing"

	"github.com/gabrie30/ghorg/configs"
)

func TestDefaultBranch(t *testing.T) {
	configs.Load()
	if os.Getenv("GHORG_BRANCH") != "master" {
		t.Errorf("Default branch should be master")
	}
}
