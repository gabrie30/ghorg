package scm

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	ghpkg "github.com/google/go-github/v72/github"
)

const (
	// baseURLPath is a non-empty Client.BaseURL path to use during tests,
	// to ensure relative URLs are used for all endpoints.
	baseURLPath = "/api-v3"
)

func setup() (client *ghpkg.Client, mux *http.ServeMux, serverURL string, teardown func()) {
	// mux is the HTTP request multiplexer used with the test server.
	mux = http.NewServeMux()

	// We want to ensure that tests catch mistakes where the endpoint URL is
	// specified as absolute rather than relative. It only makes a difference
	// when there's a non-empty base URL path. So, use that.
	apiHandler := http.NewServeMux()
	apiHandler.Handle(baseURLPath+"/", http.StripPrefix(baseURLPath, mux))
	apiHandler.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintln(os.Stderr, "FAIL: Client.BaseURL path prefix is not preserved in the request URL:")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "\t"+req.URL.String())
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "\tDid you accidentally use an absolute endpoint URL rather than relative?")
		fmt.Fprintln(os.Stderr, "\tSee https://github.com/google/go-github/issues/752 for information.")
		http.Error(w, "Client.BaseURL path prefix is not preserved in the request URL.", http.StatusInternalServerError)
	})

	// server is a test HTTP server used to provide mock API responses.
	server := httptest.NewServer(apiHandler)

	// client is the GitHub client being tested and is
	// configured to use test server.
	client = ghpkg.NewClient(nil)
	url, _ := url.Parse(server.URL + baseURLPath + "/")
	client.BaseURL = url
	client.UploadURL = url

	return client, mux, server.URL, server.Close
}

func TestGetOrgRepos(t *testing.T) {
	client, mux, _, teardown := setup()

	github := Github{Client: client}

	defer teardown()

	mux.HandleFunc("/orgs/testorg/repos", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[
			{"id":1, "clone_url": "https://example.com", "name": "foobar1", "archived": false, "fork": false, "topics": ["a","b","c"], "ssh_url": "git://example.com"},
			{"id":2, "clone_url": "https://example.com", "name": "foobar2", "archived": false, "fork": false, "topics": ["a","b","c"], "ssh_url": "git://example.com"},
			{"id":3, "clone_url": "https://example.com", "name": "foobar3", "archived": false, "fork": false, "topics": ["a","b","c"], "ssh_url": "git://example.com"},
			{"id":4, "clone_url": "https://example.com", "name": "foobar4", "archived": false, "fork": false, "topics": ["a","b","c"], "ssh_url": "git://example.com"},
			{"id":5, "clone_url": "https://example.com", "name": "foobar5", "archived": false, "fork": false, "topics": ["a","b","c"], "ssh_url": "git://example.com"},
			{"id":6, "clone_url": "https://example.com", "name": "foobar6", "archived": false, "fork": false, "topics": ["a","b","c"], "ssh_url": "git://example.com"},
			{"id":7, "clone_url": "https://example.com", "name": "foobar7", "archived": false, "fork": false, "topics": ["a","b","c"], "ssh_url": "git://example.com"},
			{"id":8, "clone_url": "https://example.com", "name": "tp-foobar8", "archived": false, "fork": false, "topics": ["a","b","c"], "ssh_url": "git://example.com"},
			{"id":9, "clone_url": "https://example.com", "name": "tp-foobar9", "archived": false, "fork": true, "topics": ["a","b","c"], "ssh_url": "git://example.com"},
			{"id":10, "clone_url": "https://example.com", "name": "tp-foobar10", "archived": true, "fork": false, "topics": ["test-topic"], "ssh_url": "httgitps://example.com"}
			]`)
	})

	t.Run("Should return all repos", func(tt *testing.T) {

		resp, err := github.GetOrgRepos("testorg")

		if err != nil {
			t.Fatal(err)
		}

		want := 10
		got := len(resp)
		if want != got {
			tt.Errorf("Expected %v repo, got: %v", want, got)
		}

	})

	t.Run("Should skip archived repos when env is set", func(tt *testing.T) {
		os.Setenv("GHORG_SKIP_ARCHIVED", "true")
		resp, err := github.GetOrgRepos("testorg")

		if err != nil {
			t.Fatal(err)
		}
		want := 9
		got := len(resp)
		if want != got {
			tt.Errorf("Expected %v repo, got: %v", want, got)
		}
		os.Setenv("GHORG_SKIP_ARCHIVED", "")

	})

	t.Run("Should skip forked repos when env is set", func(tt *testing.T) {
		os.Setenv("GHORG_SKIP_FORKS", "true")
		resp, err := github.GetOrgRepos("testorg")

		if err != nil {
			t.Fatal(err)
		}
		want := 9
		got := len(resp)
		if want != got {
			tt.Errorf("Expected %v repo, got: %v", want, got)
		}
		os.Setenv("GHORG_SKIP_FORKS", "")

	})

	t.Run("Find all repos with specific topic set", func(tt *testing.T) {
		os.Setenv("GHORG_TOPICS", "test-topic")
		resp, err := github.GetOrgRepos("testorg")

		if err != nil {
			t.Fatal(err)
		}
		want := 1
		got := len(resp)
		if want != got {
			tt.Errorf("Expected %v repo, got: %v", want, got)
		}
		os.Setenv("GHORG_TOPICS", "")
	})
}

func TestGetUserGists(t *testing.T) {
	client, mux, _, teardown := setup()

	github := Github{Client: client}

	defer teardown()

	mux.HandleFunc("/users/testuser/gists", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[
			{"id":"abc123", "git_pull_url": "https://gist.github.com/abc123.git", "public": true, "files": {"foobar.md": {"filename": "foobar.md"}}},
			{"id":"def456", "git_pull_url": "https://gist.github.com/def456.git", "public": true, "files": {"script.sh": {"filename": "script.sh"}}},
			{"id":"ghi789", "git_pull_url": "https://gist.github.com/ghi789.git", "public": false, "files": {"README.txt": {"filename": "README.txt"}}}
		]`)
	})

	t.Run("Should return all gists with HTTPS protocol", func(tt *testing.T) {
		os.Setenv("GHORG_CLONE_PROTOCOL", "https")
		os.Setenv("GHORG_GITHUB_TOKEN", "testtoken")

		resp, err := github.GetUserGists("testuser")

		if err != nil {
			tt.Fatal(err)
		}

		want := 3
		got := len(resp)
		if want != got {
			tt.Errorf("Expected %v gists, got: %v", want, got)
		}

		for _, repo := range resp {
			if !repo.IsGitHubGist {
				tt.Errorf("Expected IsGitHubGist to be true for gist %s", repo.Name)
			}
		}

		os.Unsetenv("GHORG_CLONE_PROTOCOL")
		os.Unsetenv("GHORG_GITHUB_TOKEN")
	})

	t.Run("Should use filename without extension as folder name", func(tt *testing.T) {
		os.Setenv("GHORG_CLONE_PROTOCOL", "https")
		os.Setenv("GHORG_GITHUB_TOKEN", "testtoken")

		resp, err := github.GetUserGists("testuser")
		if err != nil {
			tt.Fatal(err)
		}

		// Build a map of ID→Name for easy lookup
		names := make(map[string]string)
		for _, repo := range resp {
			names[repo.ID] = repo.Name
		}

		tests := []struct{ id, wantName string }{
			{"abc123", "foobar"}, // foobar.md  → foobar
			{"def456", "script"}, // script.sh  → script
			{"ghi789", "readme"}, // README.txt → readme (lowercased)
		}
		for _, tc := range tests {
			if got := names[tc.id]; got != tc.wantName {
				tt.Errorf("gist %s: expected folder name %q, got %q", tc.id, tc.wantName, got)
			}
		}

		os.Unsetenv("GHORG_CLONE_PROTOCOL")
		os.Unsetenv("GHORG_GITHUB_TOKEN")
	})

	t.Run("Should set correct clone branch for gists", func(tt *testing.T) {
		os.Setenv("GHORG_CLONE_PROTOCOL", "https")
		os.Setenv("GHORG_GITHUB_TOKEN", "testtoken")

		resp, err := github.GetUserGists("testuser")

		if err != nil {
			tt.Fatal(err)
		}

		// When GHORG_BRANCH is not set, gists default to "master"
		for _, repo := range resp {
			if repo.CloneBranch != "master" {
				tt.Errorf("Expected CloneBranch to be 'master' when GHORG_BRANCH is unset, got: %s", repo.CloneBranch)
			}
		}

		os.Unsetenv("GHORG_CLONE_PROTOCOL")
		os.Unsetenv("GHORG_GITHUB_TOKEN")
	})

	t.Run("Should respect GHORG_BRANCH for gists", func(tt *testing.T) {
		os.Setenv("GHORG_CLONE_PROTOCOL", "https")
		os.Setenv("GHORG_GITHUB_TOKEN", "testtoken")
		os.Setenv("GHORG_BRANCH", "main")

		resp, err := github.GetUserGists("testuser")

		if err != nil {
			tt.Fatal(err)
		}

		for _, repo := range resp {
			if repo.CloneBranch != "main" {
				tt.Errorf("Expected CloneBranch to be 'main', got: %s", repo.CloneBranch)
			}
		}

		os.Unsetenv("GHORG_CLONE_PROTOCOL")
		os.Unsetenv("GHORG_GITHUB_TOKEN")
		os.Unsetenv("GHORG_BRANCH")
	})
}

func TestGistFolderName(t *testing.T) {
	strPtr := func(s string) *string { return &s }

	tests := []struct {
		name     string
		gist     *ghpkg.Gist
		wantName string
	}{
		{
			name:     "single file with extension",
			gist:     &ghpkg.Gist{ID: strPtr("id1"), Files: map[ghpkg.GistFilename]ghpkg.GistFile{"foobar.md": {}}},
			wantName: "foobar",
		},
		{
			name:     "uppercase extension is lowercased",
			gist:     &ghpkg.Gist{ID: strPtr("id2"), Files: map[ghpkg.GistFilename]ghpkg.GistFile{"README.txt": {}}},
			wantName: "readme",
		},
		{
			name:     "no extension",
			gist:     &ghpkg.Gist{ID: strPtr("id3"), Files: map[ghpkg.GistFilename]ghpkg.GistFile{"Makefile": {}}},
			wantName: "makefile",
		},
		{
			name:     "multiple files uses first alphabetically",
			gist:     &ghpkg.Gist{ID: strPtr("id4"), Files: map[ghpkg.GistFilename]ghpkg.GistFile{"zebra.go": {}, "alpha.go": {}, "middle.go": {}}},
			wantName: "alpha",
		},
		{
			name:     "no files falls back to gist id",
			gist:     &ghpkg.Gist{ID: strPtr("abc999"), Files: map[ghpkg.GistFilename]ghpkg.GistFile{}},
			wantName: "abc999",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			got := gistFolderName(tc.gist)
			if got != tc.wantName {
				tt.Errorf("gistFolderName: expected %q, got %q", tc.wantName, got)
			}
		})
	}
}

func TestFilterGistsCollisions(t *testing.T) {
	client, mux, _, teardown := setup()
	gh := Github{Client: client}
	defer teardown()

	// Two gists share the same derived folder name ("foobar")
	mux.HandleFunc("/users/collisionuser/gists", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[
			{"id":"aaa111", "git_pull_url": "https://gist.github.com/aaa111.git", "files": {"foobar.md": {"filename": "foobar.md"}}},
			{"id":"bbb222", "git_pull_url": "https://gist.github.com/bbb222.git", "files": {"foobar.sh": {"filename": "foobar.sh"}}},
			{"id":"ccc333", "git_pull_url": "https://gist.github.com/ccc333.git", "files": {"unique.py": {"filename": "unique.py"}}}
		]`)
	})

	os.Setenv("GHORG_CLONE_PROTOCOL", "https")
	os.Setenv("GHORG_GITHUB_TOKEN", "testtoken")
	defer func() {
		os.Unsetenv("GHORG_CLONE_PROTOCOL")
		os.Unsetenv("GHORG_GITHUB_TOKEN")
	}()

	resp, err := gh.GetUserGists("collisionuser")
	if err != nil {
		t.Fatal(err)
	}

	if len(resp) != 3 {
		t.Fatalf("Expected 3 repos, got %d", len(resp))
	}

	names := make(map[string]string) // id → Name
	for _, repo := range resp {
		names[repo.ID] = repo.Name
	}

	// Both "foobar.md" and "foobar.sh" collide on "foobar", so both get the gist ID appended
	if got := names["aaa111"]; got != "foobar-aaa111" {
		t.Errorf("aaa111: expected 'foobar-aaa111', got %q", got)
	}
	if got := names["bbb222"]; got != "foobar-bbb222" {
		t.Errorf("bbb222: expected 'foobar-bbb222', got %q", got)
	}
	// No collision for "unique"
	if got := names["ccc333"]; got != "unique" {
		t.Errorf("ccc333: expected 'unique', got %q", got)
	}
}
