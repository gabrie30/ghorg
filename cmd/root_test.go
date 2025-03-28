package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/gabrie30/ghorg/git"
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

func TestGetOrSetDefaults_GHORG_USE_GIT_CLI(t *testing.T) {
	tests := []struct {
		name             string
		envValue         string
		expectedValue    string
		shouldSetToFalse bool
	}{
		{
			name:             "Default value when not set",
			envValue:         "",
			expectedValue:    "false",
			shouldSetToFalse: true,
		},
		{
			name:             "Preserve true value",
			envValue:         "true",
			expectedValue:    "true",
			shouldSetToFalse: false,
		},
		{
			name:             "Preserve false value",
			envValue:         "false",
			expectedValue:    "false",
			shouldSetToFalse: false,
		},
		{
			name:             "Preserve 1 value",
			envValue:         "1",
			expectedValue:    "1",
			shouldSetToFalse: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment
			os.Unsetenv("GHORG_USE_GIT_CLI")

			// Set test value if provided
			if tt.envValue != "" {
				os.Setenv("GHORG_USE_GIT_CLI", tt.envValue)
			}

			// Call getOrSetDefaults
			getOrSetDefaults("GHORG_USE_GIT_CLI")

			// Check result
			actualValue := os.Getenv("GHORG_USE_GIT_CLI")
			if actualValue != tt.expectedValue {
				t.Errorf("Expected %s, got %s", tt.expectedValue, actualValue)
			}

			// Cleanup
			os.Unsetenv("GHORG_USE_GIT_CLI")
		})
	}
}

func TestUseGitCLIFlagDefinition(t *testing.T) {
	// Test that the flag is properly defined
	flag := cloneCmd.Flags().Lookup("use-git-cli")
	if flag == nil {
		t.Fatal("use-git-cli flag should be defined")
	}

	// Test flag properties
	if flag.Name != "use-git-cli" {
		t.Errorf("Expected flag name 'use-git-cli', got '%s'", flag.Name)
	}

	if flag.DefValue != "false" {
		t.Errorf("Expected default value 'false', got '%s'", flag.DefValue)
	}

	expectedUsage := "GHORG_USE_GIT_CLI - Use the git CLI for cloning instead of the internal git library. This is useful for debugging and testing purposes."
	if flag.Usage != expectedUsage {
		t.Errorf("Expected usage '%s', got '%s'", expectedUsage, flag.Usage)
	}

	// Test that it's a bool flag
	if flag.Value.Type() != "bool" {
		t.Errorf("Expected bool flag, got %s", flag.Value.Type())
	}
}

func TestGHORG_USE_GIT_CLI_Integration(t *testing.T) {
	tests := []struct {
		name               string
		envValue           string
		expectedClientType string
	}{
		{
			name:               "Environment true selects CLI client",
			envValue:           "true",
			expectedClientType: "*git.GitClient",
		},
		{
			name:               "Environment false selects Go-git client",
			envValue:           "false",
			expectedClientType: "*git.GoGitClient",
		},
		{
			name:               "Environment 1 selects CLI client",
			envValue:           "1",
			expectedClientType: "*git.GitClient",
		},
		{
			name:               "Environment yes selects CLI client",
			envValue:           "yes",
			expectedClientType: "*git.GitClient",
		},
		{
			name:               "Environment empty defaults to Go-git client",
			envValue:           "",
			expectedClientType: "*git.GoGitClient",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalEnv := os.Getenv("GHORG_USE_GIT_CLI")

			// Clean up function
			defer func() {
				// Restore original environment
				if originalEnv != "" {
					os.Setenv("GHORG_USE_GIT_CLI", originalEnv)
				} else {
					os.Unsetenv("GHORG_USE_GIT_CLI")
				}
			}()

			// Set test environment
			if tt.envValue != "" {
				os.Setenv("GHORG_USE_GIT_CLI", tt.envValue)
			} else {
				os.Unsetenv("GHORG_USE_GIT_CLI")
			}

			// Simulate the logic from cloneFunc to determine useGitCLI
			// This is the actual logic used in production
			useGitCLI := false
			envValue := strings.ToLower(os.Getenv("GHORG_USE_GIT_CLI"))
			if envValue == "true" || envValue == "1" || envValue == "yes" {
				useGitCLI = true
			}

			// Create the git client using the determined value
			client := git.NewGit(useGitCLI)

			// Check the type
			clientType := ""
			switch client.(type) {
			case *git.GitClient:
				clientType = "*git.GitClient"
			case *git.GoGitClient:
				clientType = "*git.GoGitClient"
			default:
				clientType = "unknown"
			}

			if clientType != tt.expectedClientType {
				t.Errorf("With GHORG_USE_GIT_CLI=%s, expected %s, got %s",
					tt.envValue, tt.expectedClientType, clientType)
			}
		})
	}
}

func TestGHORG_USE_GIT_CLI_DefaultConfiguration(t *testing.T) {
	// Save original environment
	originalEnv := os.Getenv("GHORG_USE_GIT_CLI")

	// Clean up function
	defer func() {
		// Restore original environment
		if originalEnv != "" {
			os.Setenv("GHORG_USE_GIT_CLI", originalEnv)
		} else {
			os.Unsetenv("GHORG_USE_GIT_CLI")
		}
	}()

	// Clean environment to test default behavior
	os.Unsetenv("GHORG_USE_GIT_CLI")

	// Test the default behavior - when GHORG_USE_GIT_CLI is not set,
	// the system should default to Go-git client (useGitCLI = false)
	useGitCLI := false
	envValue := strings.ToLower(os.Getenv("GHORG_USE_GIT_CLI"))
	if envValue == "true" || envValue == "1" || envValue == "yes" {
		useGitCLI = true
	}

	// Verify useGitCLI is false by default
	if useGitCLI {
		t.Error("Default behavior should result in useGitCLI=false")
	}

	// Verify that this results in the Go-git client
	client := git.NewGit(useGitCLI)
	if _, ok := client.(*git.GoGitClient); !ok {
		t.Error("Default configuration should result in GoGitClient")
	}
}
