package configs

import (
	"os"
	"path/filepath"
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

func TestGetTokenFromFileNonExistent(t *testing.T) {
	// This test verifies that the function handles non-existent files properly
	// Note: The current implementation calls log.Fatal, so we can't easily test this
	// without changing the implementation. This is a design consideration for future improvement.
	t.Skip("Skipping test for non-existent file as current implementation calls log.Fatal")
}

