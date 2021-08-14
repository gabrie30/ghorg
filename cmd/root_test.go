package cmd

import (
	"os"
	"testing"
)

func TestDefaultSettings(t *testing.T) {
	Execute()
	protocol := os.Getenv("GHORG_CLONE_PROTOCOL")
	scm := os.Getenv("GHORG_SCM_TYPE")
	cloneType := os.Getenv("GHORG_CLONE_TYPE")

	if protocol != "https" {
		t.Errorf("Default protocol should be https, got: %v", protocol)
	}

	if scm != "github" {
		t.Errorf("Default scm should be github, got: %v", scm)
	}

	if cloneType != "org" {
		t.Errorf("Default clone type should be org, got: %v", cloneType)
	}

}
