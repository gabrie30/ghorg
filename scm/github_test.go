package scm

import (
	"fmt"
	"github.com/gabrie30/ghorg/configs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	ghpkg "github.com/google/go-github/v32/github"
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
	config := &configs.Config{}

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
		resp, err := github.GetOrgRepos(config, "testorg")

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
		config.SkipArchived = true
		resp, err := github.GetOrgRepos(config, "testorg")

		if err != nil {
			t.Fatal(err)
		}
		want := 9
		got := len(resp)
		if want != got {
			tt.Errorf("Expected %v repo, got: %v", want, got)
		}
		config.SkipArchived = false
	})

	t.Run("Should skip forked repos when env is set", func(tt *testing.T) {
		config.SkipForks = true
		resp, err := github.GetOrgRepos(config, "testorg")

		if err != nil {
			t.Fatal(err)
		}
		want := 9
		got := len(resp)
		if want != got {
			tt.Errorf("Expected %v repo, got: %v", want, got)
		}
		config.SkipForks = false
	})

	t.Run("Find all repos with specific topic set", func(tt *testing.T) {
		config.Topics = []string{"test-topic"}
		resp, err := github.GetOrgRepos(config, "testorg")

		if err != nil {
			t.Fatal(err)
		}
		want := 1
		got := len(resp)
		if want != got {
			tt.Errorf("Expected %v repo, got: %v", want, got)
		}
		config.Topics = nil
	})

	t.Run("Find all repos with specific prefix", func(tt *testing.T) {
		config.MatchPrefix = "tp-"
		resp, err := github.GetOrgRepos(config, "testorg")

		if err != nil {
			t.Fatal(err)
		}
		want := 3
		got := len(resp)
		if want != got {
			tt.Errorf("Expected %v repo, got: %v", want, got)
		}
		config.MatchPrefix = ""
	})
}
