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
