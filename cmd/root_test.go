package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultSettings(t *testing.T) {
	// Save original environment
	origProtocol := os.Getenv("GHORG_CLONE_PROTOCOL")
	origScm := os.Getenv("GHORG_SCM_TYPE")
	origCloneType := os.Getenv("GHORG_CLONE_TYPE")
	origConfig := os.Getenv("GHORG_CONFIG")

	// Create a temporary empty config file to test defaults
	tmpDir := t.TempDir()
	tmpConfig := filepath.Join(tmpDir, "test-conf.yaml")
	if err := os.WriteFile(tmpConfig, []byte("# Empty config for testing defaults\n"), 0644); err != nil {
		t.Fatalf("Failed to create temp config: %v", err)
	}

	// Clear environment variables and use empty config to test defaults
	os.Unsetenv("GHORG_CLONE_PROTOCOL")
	os.Unsetenv("GHORG_SCM_TYPE")
	os.Unsetenv("GHORG_CLONE_TYPE")
	os.Setenv("GHORG_CONFIG", tmpConfig)

	// Restore environment after test
	defer func() {
		if origProtocol != "" {
			os.Setenv("GHORG_CLONE_PROTOCOL", origProtocol)
		} else {
			os.Unsetenv("GHORG_CLONE_PROTOCOL")
		}
		if origScm != "" {
			os.Setenv("GHORG_SCM_TYPE", origScm)
		} else {
			os.Unsetenv("GHORG_SCM_TYPE")
		}
		if origCloneType != "" {
			os.Setenv("GHORG_CLONE_TYPE", origCloneType)
		} else {
			os.Unsetenv("GHORG_CLONE_TYPE")
		}
		if origConfig != "" {
			os.Setenv("GHORG_CONFIG", origConfig)
		} else {
			os.Unsetenv("GHORG_CONFIG")
		}
	}()

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
