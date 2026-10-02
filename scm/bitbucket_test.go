package scm

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ktrysmt/go-bitbucket"
)

func TestFilterServerRepos_SCMData(t *testing.T) {
	t.Setenv("GHORG_CLONE_PROTOCOL", "ssh")

	// scmId, public, and state are not parsed by ghorg but should still reach the hook
	body := `{"values":[{"slug":"repo1","id":1,"name":"repo1","scmId":"git","state":"AVAILABLE","public":true,"project":{"key":"PROJ"},"links":{"clone":[{"href":"ssh://git@bitbucket.example.com:7999/proj/repo1.git","name":"ssh"}]}}],"size":1,"isLastPage":true,"start":0}`

	t.Run("Should not include scm data by default", func(tt *testing.T) {
		var response ServerProjectResponse
		if err := json.Unmarshal([]byte(body), &response); err != nil {
			tt.Fatal(err)
		}
		repos := Bitbucket{}.filterServerRepos(response.Values)
		if len(repos) != 1 || repos[0].SCMData != nil {
			tt.Errorf("Expected 1 repo without scm data, got %+v", repos)
		}
	})

	t.Run("Should include the raw repo object when enabled", func(tt *testing.T) {
		enableSCMData(tt)
		var response ServerProjectResponse
		if err := json.Unmarshal([]byte(body), &response); err != nil {
			tt.Fatal(err)
		}
		repos := Bitbucket{}.filterServerRepos(response.Values)
		if len(repos) != 1 {
			tt.Fatalf("Expected 1 repo, got %d", len(repos))
		}
		data := scmDataMap(tt, repos[0])
		if data["scmId"] != "git" || data["public"] != true || data["state"] != "AVAILABLE" {
			tt.Errorf("Expected the Bitbucket Server repo object, got %s", repos[0].SCMData)
		}
	})
}

// Bitbucket Cloud intentionally has no scm data, see the note in Bitbucket.filter
func TestFilterCloudRepos_NoSCMData(t *testing.T) {
	enableSCMData(t)
	t.Setenv("GHORG_CLONE_PROTOCOL", "ssh")

	repos, err := Bitbucket{}.filter([]bitbucket.Repository{{
		Name:      "repo1",
		Full_name: "workspace/repo1",
		Links: map[string]any{"clone": []any{
			map[string]any{"href": "git@bitbucket.org:workspace/repo1.git", "name": "ssh"},
		}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 || repos[0].SCMData != nil {
		t.Errorf("Expected 1 repo without scm data, got %+v", repos)
	}
}

func TestInsertAppPasswordCredentialsIntoURL(t *testing.T) {
	// Set environment variables for the test
	_ = os.Setenv("GHORG_BITBUCKET_USERNAME", "ghorg")
	_ = os.Setenv("GHORG_BITBUCKET_APP_PASSWORD", "testpassword")

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
