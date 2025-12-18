package scm

import (
	"os"
	"testing"
)

func TestInsertAppPasswordCredentialsIntoURL(t *testing.T) {
	// Set environment variables for the test
	os.Setenv("GHORG_BITBUCKET_USERNAME", "ghorg")
	os.Setenv("GHORG_BITBUCKET_APP_PASSWORD", "testpassword")

	// Define a test URL
	testURL := "https://ghorg@bitbucket.org/foobar/testrepo.git"

	// Call the function with the test URL
	resultURL := insertAppPasswordCredentialsIntoURL(testURL)

	// Define the expected result
	expectedURL := "https://ghorg:testpassword@bitbucket.org/foobar/testrepo.git"

	// Check if the result matches the expected result
	if resultURL != expectedURL {
		t.Errorf("Expected %s, but got %s", expectedURL, resultURL)
	}
}

func TestInsertAPITokenCredentialsIntoURL(t *testing.T) {
	tests := []struct {
		name        string
		inputURL    string
		apiToken    string
		expectedURL string
	}{
		{
			name:        "URL with username",
			inputURL:    "https://ghorg@bitbucket.org/foobar/testrepo.git",
			apiToken:    "test_api_token",
			expectedURL: "https://x-bitbucket-api-token-auth:test_api_token@bitbucket.org/foobar/testrepo.git",
		},
		{
			name:        "URL without username",
			inputURL:    "https://bitbucket.org/foobar/testrepo.git",
			apiToken:    "test_api_token",
			expectedURL: "https://x-bitbucket-api-token-auth:test_api_token@bitbucket.org/foobar/testrepo.git",
		},
		{
			name:        "Non-HTTPS URL should be unchanged",
			inputURL:    "git@bitbucket.org:foobar/testrepo.git",
			apiToken:    "test_api_token",
			expectedURL: "git@bitbucket.org:foobar/testrepo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultURL := insertAPITokenCredentialsIntoURL(tt.inputURL, tt.apiToken)
			if resultURL != tt.expectedURL {
				t.Errorf("Expected %s, but got %s", tt.expectedURL, resultURL)
			}
		})
	}
}
