package scm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"code.gitea.io/sdk/gitea"
)

// mockGiteaRepository creates a mock Gitea repository for testing
func mockGiteaRepository(id int64, name string) *gitea.Repository {
	return &gitea.Repository{
		ID:            id,
		Name:          name,
		FullName:      fmt.Sprintf("test-org/%s", name),
		CloneURL:      fmt.Sprintf("https://gitea.example.com/test-org/%s.git", name),
		SSHURL:        fmt.Sprintf("git@gitea.example.com:test-org/%s.git", name),
		Private:       false,
		Fork:          false,
		Archived:      false,
		DefaultBranch: "main",
		Owner: &gitea.User{
			UserName: "test-org",
		},
	}
}

// setupGiteaTest creates a test server and Gitea client for testing
func setupGiteaTest() (client Gitea, mux *http.ServeMux, serverURL string, teardown func()) {
	// Create a test HTTP server
	mux = http.NewServeMux()

	// Mock the version endpoint that Gitea client calls during initialization
	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"version": "1.18.0"})
	})

	// Mock the settings endpoint
	mux.HandleFunc("/api/v1/settings/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"max_response_items": 50,
		})
	})

	server := httptest.NewServer(mux)

	// Create Gitea client with custom base URL
	giteaClient, err := gitea.NewClient(server.URL)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Gitea client: %v", err))
	}

	client = Gitea{
		Client:  giteaClient,
		perPage: 10, // Small page size for testing pagination
	}

	return client, mux, server.URL, server.Close
}

func TestGitea_GetOrgRepos_SinglePage(t *testing.T) {
	client, mux, _, teardown := setupGiteaTest()
	defer teardown()

	// Mock API response for single page
	repos := make([]*gitea.Repository, 5)
	for i := 0; i < 5; i++ {
		repos[i] = mockGiteaRepository(int64(i+1), fmt.Sprintf("repo-%03d", i+1))
	}

	mux.HandleFunc("/api/v1/orgs/test-org/repos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repos)
	})

	// Set required environment variables
	os.Setenv("GHORG_CLONE_PROTOCOL", "https")
	defer os.Unsetenv("GHORG_CLONE_PROTOCOL")

	result, err := client.GetOrgRepos("test-org")
	if err != nil {
		t.Fatalf("GetOrgRepos failed: %v", err)
	}

	if len(result) != 5 {
		t.Errorf("Expected 5 repositories, got %d", len(result))
	}

	// Verify repository data
	for i, repo := range result {
		expectedName := fmt.Sprintf("repo-%03d", i+1)
		if repo.Name != expectedName {
			t.Errorf("Expected repository name %s, got %s", expectedName, repo.Name)
		}
	}
}

func TestGitea_GetOrgRepos_MultiplePage_PaginationBugRegression(t *testing.T) {
	client, mux, _, teardown := setupGiteaTest()
	defer teardown()

	// This test specifically catches the pagination bug where perPage was undefined
	// and defaulted to 0, causing infinite loops or early termination

	pageRequests := 0
	totalRepos := 25 // More than perPage (10) to force pagination

	mux.HandleFunc("/api/v1/orgs/test-org/repos", func(w http.ResponseWriter, r *http.Request) {
		pageRequests++

		// Parse page parameter
		page := 1
		if pageParam := r.URL.Query().Get("page"); pageParam != "" {
			if p, err := fmt.Sscanf(pageParam, "%d", &page); p != 1 || err != nil {
				page = 1
			}
		}

		// Calculate repositories for this page
		startIdx := (page - 1) * client.perPage
		endIdx := startIdx + client.perPage
		if endIdx > totalRepos {
			endIdx = totalRepos
		}

		repos := make([]*gitea.Repository, 0, endIdx-startIdx)
		for i := startIdx; i < endIdx; i++ {
			repos = append(repos, mockGiteaRepository(int64(i+1), fmt.Sprintf("repo-%03d", i+1)))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repos)
	})

	// Set required environment variables
	os.Setenv("GHORG_CLONE_PROTOCOL", "https")
	defer os.Unsetenv("GHORG_CLONE_PROTOCOL")

	result, err := client.GetOrgRepos("test-org")
	if err != nil {
		t.Fatalf("GetOrgRepos failed: %v", err)
	}

	// Verify all repositories were fetched across multiple pages
	if len(result) != totalRepos {
		t.Errorf("Expected %d repositories, got %d", totalRepos, len(result))
	}

	// Verify pagination occurred (should need 3 requests: 10+10+5)
	expectedPages := 3
	if pageRequests != expectedPages {
		t.Errorf("Expected %d page requests, got %d", expectedPages, pageRequests)
	}

	// Verify repository data continuity across pages
	for i, repo := range result {
		expectedName := fmt.Sprintf("repo-%03d", i+1)
		if repo.Name != expectedName {
			t.Errorf("Expected repository name %s, got %s", expectedName, repo.Name)
		}
	}
}

func TestGitea_GetOrgRepos_ExactPageBoundary(t *testing.T) {
	client, mux, _, teardown := setupGiteaTest()
	defer teardown()

	// Test the exact scenario that caused the original bug:
	// When repositories count is exactly divisible by perPage
	totalRepos := 50 // Exactly 5 pages of 10 repos each

	mux.HandleFunc("/api/v1/orgs/test-org/repos", func(w http.ResponseWriter, r *http.Request) {
		page := 1
		if pageParam := r.URL.Query().Get("page"); pageParam != "" {
			fmt.Sscanf(pageParam, "%d", &page)
		}

		startIdx := (page - 1) * client.perPage
		endIdx := startIdx + client.perPage
		if endIdx > totalRepos {
			endIdx = totalRepos
		}

		repos := make([]*gitea.Repository, 0, endIdx-startIdx)
		for i := startIdx; i < endIdx; i++ {
			repos = append(repos, mockGiteaRepository(int64(i+1), fmt.Sprintf("repo-%03d", i+1)))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repos)
	})

	os.Setenv("GHORG_CLONE_PROTOCOL", "https")
	defer os.Unsetenv("GHORG_CLONE_PROTOCOL")

	result, err := client.GetOrgRepos("test-org")
	if err != nil {
		t.Fatalf("GetOrgRepos failed: %v", err)
	}

	// This would fail with the original bug where perPage was undefined (defaulting to 0)
	// The comparison len(rps) < perPage would be len(rps) < 0, always false, causing infinite loop
	if len(result) != totalRepos {
		t.Errorf("Expected %d repositories, got %d - this indicates a pagination bug", totalRepos, len(result))
	}
}

func TestGitea_GetUserRepos_PaginationBugRegression(t *testing.T) {
	client, mux, _, teardown := setupGiteaTest()
	defer teardown()

	// Test user repository pagination - same bug affects GetUserRepos
	totalRepos := 15

	mux.HandleFunc("/api/v1/users/test-user/repos", func(w http.ResponseWriter, r *http.Request) {
		page := 1
		if pageParam := r.URL.Query().Get("page"); pageParam != "" {
			fmt.Sscanf(pageParam, "%d", &page)
		}

		startIdx := (page - 1) * client.perPage
		endIdx := startIdx + client.perPage
		if endIdx > totalRepos {
			endIdx = totalRepos
		}

		repos := make([]*gitea.Repository, 0, endIdx-startIdx)
		for i := startIdx; i < endIdx; i++ {
			repos = append(repos, mockGiteaRepository(int64(i+1), fmt.Sprintf("user-repo-%03d", i+1)))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repos)
	})

	os.Setenv("GHORG_CLONE_PROTOCOL", "https")
	defer os.Unsetenv("GHORG_CLONE_PROTOCOL")

	result, err := client.GetUserRepos("test-user")
	if err != nil {
		t.Fatalf("GetUserRepos failed: %v", err)
	}

	if len(result) != totalRepos {
		t.Errorf("Expected %d repositories, got %d", totalRepos, len(result))
	}
}

func TestGitea_GetOrgRepos_EmptyResponse(t *testing.T) {
	client, mux, _, teardown := setupGiteaTest()
	defer teardown()

	mux.HandleFunc("/api/v1/orgs/empty-org/repos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]*gitea.Repository{})
	})

	os.Setenv("GHORG_CLONE_PROTOCOL", "https")
	defer os.Unsetenv("GHORG_CLONE_PROTOCOL")

	result, err := client.GetOrgRepos("empty-org")
	if err != nil {
		t.Fatalf("GetOrgRepos failed: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 repositories for empty org, got %d", len(result))
	}
}

func TestGitea_GetOrgRepos_NotFound(t *testing.T) {
	client, mux, _, teardown := setupGiteaTest()
	defer teardown()

	mux.HandleFunc("/api/v1/orgs/nonexistent-org/repos", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 Not Found"))
	})

	_, err := client.GetOrgRepos("nonexistent-org")
	if err == nil {
		t.Fatal("Expected error for nonexistent org, got nil")
	}

	expectedError := `org "nonexistent-org" not found`
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// Benchmark test to ensure pagination doesn't cause performance issues
func BenchmarkGitea_GetOrgRepos_LargePagination(b *testing.B) {
	client, mux, _, teardown := setupGiteaTest()
	defer teardown()

	totalRepos := 500 // Large number of repositories

	mux.HandleFunc("/api/v1/orgs/large-org/repos", func(w http.ResponseWriter, r *http.Request) {
		page := 1
		if pageParam := r.URL.Query().Get("page"); pageParam != "" {
			fmt.Sscanf(pageParam, "%d", &page)
		}

		startIdx := (page - 1) * client.perPage
		endIdx := startIdx + client.perPage
		if endIdx > totalRepos {
			endIdx = totalRepos
		}

		repos := make([]*gitea.Repository, 0, endIdx-startIdx)
		for i := startIdx; i < endIdx; i++ {
			repos = append(repos, mockGiteaRepository(int64(i+1), fmt.Sprintf("bench-repo-%03d", i+1)))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(repos)
	})

	os.Setenv("GHORG_CLONE_PROTOCOL", "https")
	defer os.Unsetenv("GHORG_CLONE_PROTOCOL")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := client.GetOrgRepos("large-org")
		if err != nil {
			b.Fatalf("GetOrgRepos failed: %v", err)
		}
		if len(result) != totalRepos {
			b.Errorf("Expected %d repositories, got %d", totalRepos, len(result))
		}
	}
}
