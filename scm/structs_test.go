package scm

import (
	"encoding/json"
	"reflect"
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
