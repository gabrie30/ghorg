package scm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"code.gitea.io/sdk/gitea"
)

func TestCodeberg_GetType(t *testing.T) {
	if got := (Codeberg{}).GetType(); got != "codeberg" {
		t.Errorf("Expected type 'codeberg', got '%s'", got)
	}
}

// TestNewGiteaClient_ThreadsProvidedToken verifies that the shared Gitea backend
// uses the token it is given (rather than reading GHORG_GITEA_TOKEN directly).
// This is what allows the Codeberg scm to authenticate with its own
// GHORG_CODEBERG_TOKEN while reusing the Gitea client.
func TestNewGiteaClient_ThreadsProvidedToken(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"version": "1.18.0"})
	})
	mux.HandleFunc("/api/v1/settings/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"max_response_items": 50})
	})

	privateRepo := mockGiteaRepository(1, "private-repo")
	privateRepo.Private = true
	mux.HandleFunc("/api/v1/orgs/test-org/repos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]*gitea.Repository{privateRepo})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Ensure the token comes from the argument, not the gitea env var.
	t.Setenv("GHORG_GITEA_TOKEN", "")
	t.Setenv("GHORG_CLONE_PROTOCOL", "https")

	// insecure=true because httptest serves over http.
	c, err := newGiteaClient(server.URL, "provided-token", true, "--insecure-gitea-client")
	if err != nil {
		t.Fatalf("newGiteaClient failed: %v", err)
	}

	result, err := c.GetOrgRepos("test-org")
	if err != nil {
		t.Fatalf("GetOrgRepos failed: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 repository, got %d", len(result))
	}

	if !strings.Contains(result[0].CloneURL, "provided-token@") {
		t.Errorf("Expected clone URL to embed the provided token, got %q", result[0].CloneURL)
	}
}

// TestCodeberg_NewClient_UsesCodebergToken verifies the Codeberg client reads
// GHORG_CODEBERG_TOKEN and threads it through to clone URLs. A mock TLS server
// stands in for codeberg.org so the https-only client can be exercised without
// network access.
func TestCodeberg_NewClient_UsesCodebergToken(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"version": "1.18.0"})
	})
	mux.HandleFunc("/api/v1/settings/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"max_response_items": 50})
	})

	privateRepo := mockGiteaRepository(1, "private-repo")
	privateRepo.Private = true
	mux.HandleFunc("/api/v1/orgs/test-org/repos", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]*gitea.Repository{privateRepo})
	})

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	// Trust the mock server's self-signed cert for the duration of the test so
	// the https-only Codeberg client can connect.
	oldTransport := http.DefaultTransport
	http.DefaultTransport = server.Client().Transport
	defer func() { http.DefaultTransport = oldTransport }()

	t.Setenv("GHORG_SCM_BASE_URL", server.URL)
	t.Setenv("GHORG_CODEBERG_TOKEN", "codeberg-token")
	t.Setenv("GHORG_CLONE_PROTOCOL", "https")

	c, err := Codeberg{}.NewClient()
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if c.GetType() != "codeberg" {
		t.Errorf("Expected client type 'codeberg', got '%s'", c.GetType())
	}

	result, err := c.GetOrgRepos("test-org")
	if err != nil {
		t.Fatalf("GetOrgRepos failed: %v", err)
	}

	if len(result) != 1 || !strings.Contains(result[0].CloneURL, "codeberg-token@") {
		t.Errorf("Expected clone URL to embed GHORG_CODEBERG_TOKEN, got %+v", result)
	}
}

// TestCodeberg_NewClient_InsecureClient verifies that a self-hosted Forgejo
// instance served over http can be used when GHORG_INSECURE_CODEBERG_CLIENT is
// set.
func TestCodeberg_NewClient_InsecureClient(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/version", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"version": "1.18.0"})
	})
	mux.HandleFunc("/api/v1/settings/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"max_response_items": 50})
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	t.Setenv("GHORG_SCM_BASE_URL", server.URL)
	t.Setenv("GHORG_CODEBERG_TOKEN", "codeberg-token")
	t.Setenv("GHORG_INSECURE_CODEBERG_CLIENT", "true")

	c, err := Codeberg{}.NewClient()
	if err != nil {
		t.Fatalf("NewClient failed with insecure client enabled: %v", err)
	}

	if c.GetType() != "codeberg" {
		t.Errorf("Expected client type 'codeberg', got '%s'", c.GetType())
	}
}
