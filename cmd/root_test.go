package cmd

import (
	"os"
	"path/filepath"
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

func TestConfigFileValuesAreExportedToEnv(t *testing.T) {
	confFile := filepath.Join(t.TempDir(), "conf.yaml")
	conf := `GHORG_PROTECT_LOCAL: true
GHORG_FETCH_GIT_LFS: true
GHORG_FETCH_PRUNE: true
GHORG_ONLY_PATH: /custom/path/to/ghorgonly
`
	if err := os.WriteFile(confFile, []byte(conf), 0o644); err != nil {
		t.Fatal(err)
	}

	expected := map[string]string{
		"GHORG_PROTECT_LOCAL": "true",
		"GHORG_FETCH_GIT_LFS": "true",
		"GHORG_FETCH_PRUNE":   "true",
		"GHORG_ONLY_PATH":     "/custom/path/to/ghorgonly",
	}

	// Clear any values leaked from other tests so viper's AutomaticEnv
	// doesn't shadow the config file. t.Setenv registers the restore,
	// os.Unsetenv removes the empty value viper would otherwise read.
	for envVar := range expected {
		t.Setenv(envVar, "")
		_ = os.Unsetenv(envVar)
	}

	// InitConfig prefers the package-level config var (set by --config or a
	// previous Execute run) over GHORG_CONFIG, so clear it for this test.
	originalConfig := config
	config = ""
	defer func() { config = originalConfig }()

	t.Setenv("GHORG_CONFIG", confFile)
	InitConfig()

	for envVar, want := range expected {
		if got := os.Getenv(envVar); got != want {
			t.Errorf("%s from config file should be exported as %q, got: %q", envVar, want, got)
		}
	}
}
