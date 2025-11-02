package scm

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func setupSourcehut() (client Sourcehut, mux *http.ServeMux, serverURL string, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)

	client = Sourcehut{
		Client:  &http.Client{},
		Token:   "test-token",
		BaseURL: server.URL,
	}

	return client, mux, server.URL, server.Close
}

func TestSourcehutGetUserRepos(t *testing.T) {
	client, mux, _, teardown := setupSourcehut()
	defer teardown()

	mux.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and headers
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "bearer test-token" {
			t.Errorf("Expected bearer token authorization")
		}

		// Return mock GraphQL response (note: API returns usernames with ~ prefix)
		fmt.Fprint(w, `{
			"data": {
				"repositories": {
					"results": [
						{
							"id": 1,
							"name": "repo1",
							"visibility": "PUBLIC",
							"owner": {"canonicalName": "~testuser"},
							"HEAD": {"name": "refs/heads/main"}
						},
						{
							"id": 2,
							"name": "repo2",
							"visibility": "PRIVATE",
							"owner": {"canonicalName": "~testuser"},
							"HEAD": {"name": "refs/heads/master"}
						},
						{
							"id": 3,
							"name": "repo3",
							"visibility": "UNLISTED",
							"owner": {"canonicalName": "~testuser"},
							"HEAD": {"name": "refs/heads/develop"}
						}
					],
					"cursor": ""
				}
			}
		}`)
	})

	t.Run("Should return all repos", func(tt *testing.T) {
		repos, err := client.GetUserRepos("testuser")
		if err != nil {
			tt.Fatal(err)
		}

		want := 3
		got := len(repos)
		if want != got {
			tt.Errorf("Expected %v repos, got: %v", want, got)
		}

		// Verify first repo details
		if repos[0].Name != "repo1" {
			tt.Errorf("Expected repo name 'repo1', got: %v", repos[0].Name)
		}
		if repos[0].CloneBranch != "main" {
			tt.Errorf("Expected clone branch 'main', got: %v", repos[0].CloneBranch)
		}
		// Path should NOT have ~ prefix
		if repos[0].Path != "testuser/repo1" {
			tt.Errorf("Expected path 'testuser/repo1', got: %v", repos[0].Path)
		}
	})

	t.Run("Should use HTTPS protocol by default", func(tt *testing.T) {
		os.Setenv("GHORG_CLONE_PROTOCOL", "")
		repos, err := client.GetUserRepos("testuser")
		if err != nil {
			tt.Fatal(err)
		}

		// All repos should use HTTPS URLs
		for _, repo := range repos {
			u, err := url.Parse(repo.CloneURL)
			if err != nil {
				tt.Errorf("Invalid clone URL: %v", repo.CloneURL)
			}
			if u.Scheme != "http" && u.Scheme != "https" {
				tt.Errorf("Expected HTTP(S) protocol, got: %v", u.Scheme)
			}
		}
	})

	t.Run("Should use SSH protocol when specified", func(tt *testing.T) {
		os.Setenv("GHORG_CLONE_PROTOCOL", "ssh")
		repos, err := client.GetUserRepos("testuser")
		if err != nil {
			tt.Fatal(err)
		}

		// All repos should use SSH URLs (git@...)
		for _, repo := range repos {
			if len(repo.CloneURL) < 4 || repo.CloneURL[:4] != "git@" {
				tt.Errorf("Expected SSH URL (git@...), got: %v", repo.CloneURL)
			}
		}
		os.Setenv("GHORG_CLONE_PROTOCOL", "")
	})

	t.Run("Should respect custom branch setting", func(tt *testing.T) {
		os.Setenv("GHORG_BRANCH", "custom-branch")
		repos, err := client.GetUserRepos("testuser")
		if err != nil {
			tt.Fatal(err)
		}

		for _, repo := range repos {
			if repo.CloneBranch != "custom-branch" {
				tt.Errorf("Expected clone branch 'custom-branch', got: %v", repo.CloneBranch)
			}
		}
		os.Setenv("GHORG_BRANCH", "")
	})
}

func TestSourcehutGetOrgRepos(t *testing.T) {
	client, mux, _, teardown := setupSourcehut()
	defer teardown()

	mux.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"data": {
				"repositories": {
					"results": [
						{
							"id": 1,
							"name": "org-repo1",
							"visibility": "PUBLIC",
							"owner": {"canonicalName": "~testorg"},
							"HEAD": {"name": "refs/heads/main"}
						}
					],
					"cursor": ""
				}
			}
		}`)
	})

	t.Run("Should return org repos (same as user repos)", func(tt *testing.T) {
		repos, err := client.GetOrgRepos("testorg")
		if err != nil {
			tt.Fatal(err)
		}

		want := 1
		got := len(repos)
		if want != got {
			tt.Errorf("Expected %v repo, got: %v", want, got)
		}

		if repos[0].Name != "org-repo1" {
			tt.Errorf("Expected repo name 'org-repo1', got: %v", repos[0].Name)
		}
		// Path should NOT have ~ prefix
		if repos[0].Path != "testorg/org-repo1" {
			tt.Errorf("Expected path 'testorg/org-repo1', got: %v", repos[0].Path)
		}
	})
}

func TestSourcehutPagination(t *testing.T) {
	client, mux, _, teardown := setupSourcehut()
	defer teardown()

	callCount := 0
	mux.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First page with cursor
			fmt.Fprint(w, `{
				"data": {
					"repositories": {
						"results": [
							{
								"id": 1,
								"name": "repo1",
								"visibility": "PUBLIC",
								"owner": {"canonicalName": "~testuser"},
								"HEAD": {"name": "refs/heads/main"}
							}
						],
						"cursor": "next-page-cursor"
					}
				}
			}`)
		} else {
			// Second page without cursor (last page)
			fmt.Fprint(w, `{
				"data": {
					"repositories": {
						"results": [
							{
								"id": 2,
								"name": "repo2",
								"visibility": "PUBLIC",
								"owner": {"canonicalName": "~testuser"},
								"HEAD": {"name": "refs/heads/main"}
							}
						],
						"cursor": ""
					}
				}
			}`)
		}
	})

	t.Run("Should handle pagination correctly", func(tt *testing.T) {
		repos, err := client.GetUserRepos("testuser")
		if err != nil {
			tt.Fatal(err)
		}

		// Should have made 2 API calls
		if callCount != 2 {
			tt.Errorf("Expected 2 API calls, got: %v", callCount)
		}

		// Should have 2 repos total
		want := 2
		got := len(repos)
		if want != got {
			tt.Errorf("Expected %v repos, got: %v", want, got)
		}
	})
}

func TestSourcehutErrorHandling(t *testing.T) {
	client, mux, _, teardown := setupSourcehut()
	defer teardown()

	t.Run("Should handle GraphQL errors", func(tt *testing.T) {
		mux.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, `{
				"errors": [
					{"message": "Authentication failed"}
				]
			}`)
		})

		_, err := client.GetUserRepos("testuser")
		if err == nil {
			tt.Error("Expected error for GraphQL error response")
		}
	})
}

func TestSourcehutNewClient(t *testing.T) {
	t.Run("Should use default base URL when not specified", func(tt *testing.T) {
		os.Setenv("GHORG_SCM_BASE_URL", "")
		os.Setenv("GHORG_SOURCEHUT_TOKEN", "test-token")
		os.Setenv("GHORG_INSECURE_SOURCEHUT_CLIENT", "false")

		client, err := Sourcehut{}.NewClient()
		if err != nil {
			tt.Fatal(err)
		}

		sourcehutClient := client.(Sourcehut)
		if sourcehutClient.BaseURL != "https://git.sr.ht" {
			tt.Errorf("Expected default base URL 'https://git.sr.ht', got: %v", sourcehutClient.BaseURL)
		}
	})

	t.Run("Should use custom base URL when specified", func(tt *testing.T) {
		os.Setenv("GHORG_SCM_BASE_URL", "https://git.custom.com")
		os.Setenv("GHORG_SOURCEHUT_TOKEN", "test-token")

		client, err := Sourcehut{}.NewClient()
		if err != nil {
			tt.Fatal(err)
		}

		sourcehutClient := client.(Sourcehut)
		if sourcehutClient.BaseURL != "https://git.custom.com" {
			tt.Errorf("Expected base URL 'https://git.custom.com', got: %v", sourcehutClient.BaseURL)
		}

		os.Setenv("GHORG_SCM_BASE_URL", "")
	})
}

func TestUsernameNormalization(t *testing.T) {
	t.Run("normalizeUsername should add ~ prefix if missing", func(tt *testing.T) {
		result := normalizeUsername("testuser")
		expected := "~testuser"
		if result != expected {
			tt.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("normalizeUsername should keep ~ prefix if present", func(tt *testing.T) {
		result := normalizeUsername("~testuser")
		expected := "~testuser"
		if result != expected {
			tt.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("stripUsernamePrefix should remove ~ prefix", func(tt *testing.T) {
		result := stripUsernamePrefix("~testuser")
		expected := "testuser"
		if result != expected {
			tt.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("stripUsernamePrefix should work with no ~ prefix", func(tt *testing.T) {
		result := stripUsernamePrefix("testuser")
		expected := "testuser"
		if result != expected {
			tt.Errorf("Expected '%s', got '%s'", expected, result)
		}
	})
}

func TestSourcehutUsernameHandling(t *testing.T) {
	client, mux, _, teardown := setupSourcehut()
	defer teardown()

	mux.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"data": {
				"repositories": {
					"results": [
						{
							"id": 1,
							"name": "test-repo",
							"visibility": "PUBLIC",
							"owner": {"canonicalName": "~testuser"},
							"HEAD": {"name": "refs/heads/main"}
						}
					],
					"cursor": ""
				}
			}
		}`)
	})

	t.Run("Should accept username without ~ prefix", func(tt *testing.T) {
		repos, err := client.GetUserRepos("testuser")
		if err != nil {
			tt.Fatal(err)
		}

		if len(repos) != 1 {
			tt.Fatalf("Expected 1 repo, got %d", len(repos))
		}

		// Path should NOT have ~ prefix
		if repos[0].Path != "testuser/test-repo" {
			tt.Errorf("Expected path 'testuser/test-repo', got '%s'", repos[0].Path)
		}
	})

	t.Run("Should accept username with ~ prefix", func(tt *testing.T) {
		repos, err := client.GetUserRepos("~testuser")
		if err != nil {
			tt.Fatal(err)
		}

		if len(repos) != 1 {
			tt.Fatalf("Expected 1 repo, got %d", len(repos))
		}

		// Path should NOT have ~ prefix (stripped)
		if repos[0].Path != "testuser/test-repo" {
			tt.Errorf("Expected path 'testuser/test-repo', got '%s'", repos[0].Path)
		}
	})

	t.Run("Clone URLs should include ~ prefix", func(tt *testing.T) {
		repos, err := client.GetUserRepos("testuser")
		if err != nil {
			tt.Fatal(err)
		}

		if len(repos) != 1 {
			tt.Fatalf("Expected 1 repo, got %d", len(repos))
		}

		// Clone URL MUST have ~ prefix for git to work
		expectedURL := client.BaseURL + "/~testuser/test-repo"
		if repos[0].CloneURL != expectedURL {
			tt.Errorf("Expected clone URL '%s', got '%s'", expectedURL, repos[0].CloneURL)
		}
	})
}

