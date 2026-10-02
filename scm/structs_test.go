package scm

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestRepoJSONContract(t *testing.T) {
	repo := Repo{
		ID:                       "42",
		Name:                     "ghorg",
		HostPath:                 "/tmp/ghorg",
		Path:                     "/group/ghorg",
		URL:                      "https://github.com/gabrie30/ghorg",
		CloneURL:                 "https://github.com/gabrie30/ghorg.git",
		CloneBranch:              "master",
		IsWiki:                   true,
		IsGitLabSnippet:          true,
		IsGitLabRootLevelSnippet: true,
		IsGitHubGist:             true,
		GitLabSnippetInfo: GitLabSnippet{
			ID:         "7",
			Title:      "snippet title",
			URLOfRepo:  "https://gitlab.com/group/repo",
			NameOfRepo: "repo",
		},
		Commits: RepoCommits{
			CountPrePull:  1,
			CountPostPull: 2,
			CountDiff:     1,
		},
	}

	data, err := json.Marshal(repo)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var asMap map[string]interface{}
	if err := json.Unmarshal(data, &asMap); err != nil {
		t.Fatalf("unmarshal to map failed: %v", err)
	}

	expectedKeys := []string{
		"id", "name", "host_path", "path", "url", "clone_url",
		"clone_branch", "is_wiki", "is_gitlab_snippet",
		"is_gitlab_root_level_snippet", "is_github_gist",
		"gitlab_snippet_info", "commits",
	}
	for _, key := range expectedKeys {
		if _, ok := asMap[key]; !ok {
			t.Errorf("expected JSON key %q missing from %s", key, data)
		}
	}
	if len(asMap) != len(expectedKeys) {
		t.Errorf("expected %d top level keys, got %d: %s", len(expectedKeys), len(asMap), data)
	}

	snippetKeys := asMap["gitlab_snippet_info"].(map[string]interface{})
	for _, key := range []string{"id", "title", "url_of_repo", "name_of_repo"} {
		if _, ok := snippetKeys[key]; !ok {
			t.Errorf("expected gitlab_snippet_info key %q missing from %s", key, data)
		}
	}

	commitKeys := asMap["commits"].(map[string]interface{})
	for _, key := range []string{"count_pre_pull", "count_post_pull", "count_diff"} {
		if _, ok := commitKeys[key]; !ok {
			t.Errorf("expected commits key %q missing from %s", key, data)
		}
	}
}

func TestRepoJSONRoundTrip(t *testing.T) {
	original := Repo{
		ID:          "42",
		Name:        "ghorg",
		HostPath:    "/tmp/ghorg",
		Path:        "/group/ghorg",
		URL:         "https://github.com/gabrie30/ghorg",
		CloneURL:    "https://github.com/gabrie30/ghorg.git",
		CloneBranch: "master",
		IsWiki:      true,
		GitLabSnippetInfo: GitLabSnippet{
			ID:    "7",
			Title: "snippet title",
		},
		Commits: RepoCommits{CountPrePull: 3},
		SCMData: json.RawMessage(`{"archived":true}`),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded Repo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if !reflect.DeepEqual(original, decoded) {
		t.Errorf("round trip mismatch:\noriginal: %+v\ndecoded:  %+v", original, decoded)
	}
}

func TestRepoJSONSCMData(t *testing.T) {
	data, err := json.Marshal(Repo{Name: "ghorg", SCMData: json.RawMessage(`{"archived":true}`)})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var asMap map[string]interface{}
	if err := json.Unmarshal(data, &asMap); err != nil {
		t.Fatalf("unmarshal to map failed: %v", err)
	}

	scmData, ok := asMap["scm_data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected scm_data to be a JSON object in %s", data)
	}
	if scmData["archived"] != true {
		t.Errorf("expected scm_data to pass the SCM object through unchanged, got %s", data)
	}

	// scm_data is omitted entirely when it was not requested
	data, err = json.Marshal(Repo{Name: "ghorg"})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if strings.Contains(string(data), "scm_data") {
		t.Errorf("expected no scm_data key when unset, got %s", data)
	}
}

// enableSCMData turns on scm_data, which needs both GHORG_INCLUDE_SCM_DATA and a repo filter hook
func enableSCMData(t *testing.T) {
	t.Helper()
	t.Setenv("GHORG_INCLUDE_SCM_DATA", "true")
	t.Setenv("GHORG_REPO_FILTER_HOOK", "/path/to/hook")
}

// scmDataMap decodes a repo's scm_data so tests can assert on individual fields
func scmDataMap(t *testing.T, r Repo) map[string]interface{} {
	t.Helper()
	var data map[string]interface{}
	if err := json.Unmarshal(r.SCMData, &data); err != nil {
		t.Fatalf("scm_data for %q is not a JSON object: %v, got %s", r.Name, err, r.SCMData)
	}
	return data
}

func TestSCMData(t *testing.T) {
	type apiRepo struct {
		Name     string `json:"name"`
		Archived bool   `json:"archived"`
	}
	repo := apiRepo{Name: "ghorg", Archived: true}

	t.Run("nil when the flag is not set", func(tt *testing.T) {
		tt.Setenv("GHORG_INCLUDE_SCM_DATA", "")
		tt.Setenv("GHORG_REPO_FILTER_HOOK", "/path/to/hook")
		if got := scmData(repo); got != nil {
			tt.Errorf("expected nil, got %s", got)
		}
	})

	t.Run("nil when no repo filter hook is set", func(tt *testing.T) {
		tt.Setenv("GHORG_INCLUDE_SCM_DATA", "true")
		tt.Setenv("GHORG_REPO_FILTER_HOOK", "")
		if got := scmData(repo); got != nil {
			tt.Errorf("expected nil, got %s", got)
		}
	})

	t.Run("JSON of the SCM object when the flag and a hook are set", func(tt *testing.T) {
		enableSCMData(tt)
		if got := string(scmData(repo)); got != `{"name":"ghorg","archived":true}` {
			tt.Errorf("unexpected scm_data: %s", got)
		}
	})
}
