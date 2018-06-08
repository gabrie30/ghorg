package main

import (
	"testing"

	"github.com/gabrie30/ghorg/config"
)

func TestDefaultBranch(t *testing.T) {
	branch := config.GhorgBranch
	if branch != "master" {
		t.Errorf("Default branch should be master")
	}
}
