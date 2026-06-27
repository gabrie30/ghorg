package configs

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetTokenFromFile(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "ghorg_token_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name        string
		fileContent string
		expected    string
		description string
	}{
		{
			name:        "clean_token",
			fileContent: "ghp_1234567890abcdef1234567890abcdef12345678",
			expected:    "ghp_1234567890abcdef1234567890abcdef12345678",
			description: "Clean token without any extra characters",
		},
		{
			name:        "token_with_newline",
			fileContent: "ghp_1234567890abcdef1234567890abcdef12345678\n",
			expected:    "ghp_1234567890abcdef1234567890abcdef12345678",
			description: "Token with trailing newline",
		},
		{
			name:        "token_with_carriage_return",
			fileContent: "ghp_1234567890abcdef1234567890abcdef12345678\r\n",
			expected:    "ghp_1234567890abcdef1234567890abcdef12345678",
			description: "Token with Windows-style line ending",
		},
		{
			name:        "token_with_spaces",
			fileContent: "  ghp_1234567890abcdef1234567890abcdef12345678  ",
			expected:    "ghp_1234567890abcdef1234567890abcdef12345678",
			description: "Token with leading and trailing spaces",
		},
		{
			name:        "token_with_tabs",
			fileContent: "\tghp_1234567890abcdef1234567890abcdef12345678\t",
			expected:    "ghp_1234567890abcdef1234567890abcdef12345678",
			description: "Token with leading and trailing tabs",
		},
		{
			name:        "token_with_bom",
			fileContent: "\xef\xbb\xbfghp_1234567890abcdef1234567890abcdef12345678",
			expected:    "ghp_1234567890abcdef1234567890abcdef12345678",
			description: "Token with UTF-8 BOM",
		},
		{
			name:        "token_with_mixed_whitespace",
			fileContent: "\r\n  \tghp_1234567890abcdef1234567890abcdef12345678\t  \r\n",
			expected:    "ghp_1234567890abcdef1234567890abcdef12345678",
			description: "Token with mixed whitespace and control characters",
		},
		{
			name:        "token_with_control_chars",
			fileContent: "\x00\x01ghp_1234567890abcdef1234567890abcdef12345678\x7f\x1f",
			expected:    "ghp_1234567890abcdef1234567890abcdef12345678",
			description: "Token with various control characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tempDir, tt.name+".txt")
			err := os.WriteFile(testFile, []byte(tt.fileContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Test GetTokenFromFile
			result := GetTokenFromFile(testFile)
			if result != tt.expected {
				t.Errorf("GetTokenFromFile() = %q, expected %q\nDescription: %s", result, tt.expected, tt.description)
			}

			// Verify the result contains only valid HTTP header characters
			for _, r := range result {
				if r <= 32 || r > 126 {
					t.Errorf("Result contains invalid character: %d (0x%x) in %q", r, r, result)
				}
			}
		})
	}
}

func TestRunTokenCmd(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("runTokenCmd test relies on a unix shell")
	}

	// Env vars touched by runTokenCmd, cleared before each subtest.
	tokenEnvs := []string{
		"GHORG_TOKEN_CMD",
		"GHORG_SCM_TYPE",
		"GHORG_GITHUB_TOKEN",
		"GHORG_GITLAB_TOKEN",
		"GHORG_GITEA_TOKEN",
		"GHORG_SOURCEHUT_TOKEN",
		"GHORG_BITBUCKET_USERNAME",
		"GHORG_BITBUCKET_API_EMAIL",
		"GHORG_BITBUCKET_APP_PASSWORD",
		"GHORG_BITBUCKET_API_TOKEN",
		"GHORG_BITBUCKET_OAUTH_TOKEN",
	}
	clearEnvs := func() {
		for _, e := range tokenEnvs {
			os.Unsetenv(e)
		}
	}

	tests := []struct {
		name     string
		setup    map[string]string
		checkEnv string
		expected string
	}{
		{
			name:     "github_sets_token",
			setup:    map[string]string{"GHORG_SCM_TYPE": "github", "GHORG_TOKEN_CMD": "echo gh-token"},
			checkEnv: "GHORG_GITHUB_TOKEN",
			expected: "gh-token",
		},
		{
			name:     "gitlab_sets_token",
			setup:    map[string]string{"GHORG_SCM_TYPE": "gitlab", "GHORG_TOKEN_CMD": "echo gl-token"},
			checkEnv: "GHORG_GITLAB_TOKEN",
			expected: "gl-token",
		},
		{
			name:     "output_is_trimmed",
			setup:    map[string]string{"GHORG_SCM_TYPE": "gitea", "GHORG_TOKEN_CMD": "printf '  gt-token \\n'"},
			checkEnv: "GHORG_GITEA_TOKEN",
			expected: "gt-token",
		},
		{
			name:     "existing_token_takes_precedence",
			setup:    map[string]string{"GHORG_SCM_TYPE": "github", "GHORG_GITHUB_TOKEN": "already-set", "GHORG_TOKEN_CMD": "echo from-cmd"},
			checkEnv: "GHORG_GITHUB_TOKEN",
			expected: "already-set",
		},
		{
			name:     "bitbucket_username_uses_app_password",
			setup:    map[string]string{"GHORG_SCM_TYPE": "bitbucket", "GHORG_BITBUCKET_USERNAME": "user", "GHORG_TOKEN_CMD": "echo bb-app"},
			checkEnv: "GHORG_BITBUCKET_APP_PASSWORD",
			expected: "bb-app",
		},
		{
			name:     "bitbucket_api_email_uses_api_token",
			setup:    map[string]string{"GHORG_SCM_TYPE": "bitbucket", "GHORG_BITBUCKET_API_EMAIL": "user@example.com", "GHORG_TOKEN_CMD": "echo bb-api"},
			checkEnv: "GHORG_BITBUCKET_API_TOKEN",
			expected: "bb-api",
		},
		{
			name:     "bitbucket_default_uses_oauth",
			setup:    map[string]string{"GHORG_SCM_TYPE": "bitbucket", "GHORG_TOKEN_CMD": "echo bb-oauth"},
			checkEnv: "GHORG_BITBUCKET_OAUTH_TOKEN",
			expected: "bb-oauth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnvs()
			defer clearEnvs()
			for k, v := range tt.setup {
				os.Setenv(k, v)
			}

			runTokenCmd()

			if got := os.Getenv(tt.checkEnv); got != tt.expected {
				t.Errorf("%s = %q, expected %q", tt.checkEnv, got, tt.expected)
			}
		})
	}

	t.Run("noop_when_cmd_unset", func(t *testing.T) {
		clearEnvs()
		defer clearEnvs()
		os.Setenv("GHORG_SCM_TYPE", "github")

		runTokenCmd()

		if got := os.Getenv("GHORG_GITHUB_TOKEN"); got != "" {
			t.Errorf("GHORG_GITHUB_TOKEN = %q, expected empty", got)
		}
	})
}

func TestGetTokenFromFileNonExistent(t *testing.T) {
	// This test verifies that the function handles non-existent files properly
	// Note: The current implementation calls log.Fatal, so we can't easily test this
	// without changing the implementation. This is a design consideration for future improvement.
	t.Skip("Skipping test for non-existent file as current implementation calls log.Fatal")
}
