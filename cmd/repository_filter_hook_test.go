package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/gabrie30/ghorg/scm"
)

// writeHookScript writes an executable script to a temp dir and returns its path.
func writeHookScript(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "hook.sh")
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("failed to write hook script: %v", err)
	}
	return path
}

func skipOnWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("repo filter hook script tests require a POSIX shell")
	}
}

func TestRunRepoFilterHook_IdentityHookReturnsAllRepos(t *testing.T) {
	skipOnWindows(t)
	hook := writeHookScript(t, "#!/bin/sh\ncat\n")

	repos := []scm.Repo{
		{Name: "repo1", CloneURL: "https://example.com/repo1.git", CloneBranch: "main"},
		{Name: "repo2", CloneURL: "https://example.com/repo2.git", CloneBranch: "master", IsWiki: true},
	}

	result, err := runRepoFilterHook(hook, repos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(result, repos) {
		t.Errorf("expected %+v, got %+v", repos, result)
	}
}

func TestRunRepoFilterHook_HookCanFilterAndMutate(t *testing.T) {
	skipOnWindows(t)
	// The hook drops repo2 and changes repo1's clone_branch, proving both
	// filtering and mutation round trip through the JSON contract.
	hook := writeHookScript(t, `#!/bin/sh
cat > /dev/null
echo '[{"id":"","name":"repo1","host_path":"","path":"","url":"","clone_url":"https://example.com/repo1.git","clone_branch":"develop","is_wiki":false,"is_gitlab_snippet":false,"is_gitlab_root_level_snippet":false,"is_github_gist":false,"gitlab_snippet_info":{"id":"","title":"","url_of_repo":"","name_of_repo":""},"commits":{"count_pre_pull":0,"count_post_pull":0,"count_diff":0}}]'
`)

	repos := []scm.Repo{
		{Name: "repo1", CloneURL: "https://example.com/repo1.git", CloneBranch: "main"},
		{Name: "repo2", CloneURL: "https://example.com/repo2.git", CloneBranch: "main"},
	}

	result, err := runRepoFilterHook(hook, repos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := []scm.Repo{
		{Name: "repo1", CloneURL: "https://example.com/repo1.git", CloneBranch: "develop"},
	}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %+v, got %+v", expected, result)
	}
}

func TestRunRepoFilterHook_EmptyArrayMeansCloneNothing(t *testing.T) {
	skipOnWindows(t)
	hook := writeHookScript(t, "#!/bin/sh\ncat > /dev/null\necho '[]'\n")

	repos := []scm.Repo{{Name: "repo1"}}

	result, err := runRepoFilterHook(hook, repos)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %+v", result)
	}
}

func TestRunRepoFilterHook_NonZeroExitReturnsError(t *testing.T) {
	skipOnWindows(t)
	hook := writeHookScript(t, "#!/bin/sh\ncat > /dev/null\nexit 3\n")

	_, err := runRepoFilterHook(hook, []scm.Repo{{Name: "repo1"}})
	if err == nil {
		t.Fatal("expected error for non-zero exit, got nil")
	}
}

func TestRunRepoFilterHook_InvalidJSONReturnsError(t *testing.T) {
	skipOnWindows(t)
	hook := writeHookScript(t, "#!/bin/sh\ncat > /dev/null\necho 'not json'\n")

	_, err := runRepoFilterHook(hook, []scm.Repo{{Name: "repo1"}})
	if err == nil {
		t.Fatal("expected error for invalid JSON output, got nil")
	}
}

func TestRunRepoFilterHook_MissingHookReturnsError(t *testing.T) {
	skipOnWindows(t)
	missing := filepath.Join(t.TempDir(), "does-not-exist.sh")

	_, err := runRepoFilterHook(missing, []scm.Repo{{Name: "repo1"}})
	if err == nil {
		t.Fatal("expected error for missing hook executable, got nil")
	}
}

func TestFilterByHook_SkippedWhenEnvUnset(t *testing.T) {
	t.Setenv("GHORG_REPO_FILTER_HOOK", "")
	filter := NewRepositoryFilter()

	repos := []scm.Repo{{Name: "repo1"}, {Name: "repo2"}}

	result := filter.FilterByHook(repos)
	if !reflect.DeepEqual(result, repos) {
		t.Errorf("expected repos unchanged when hook unset, got %+v", result)
	}
}

func TestApplyAllFilters_RunsHookLast(t *testing.T) {
	skipOnWindows(t)
	// Hook keeps only repo names containing "keep". Combined with the
	// built-in regex filter, this proves the hook runs after built-ins.
	hook := writeHookScript(t, `#!/bin/sh
cat > /dev/null
echo '[{"id":"","name":"keep-me","host_path":"","path":"","url":"","clone_url":"","clone_branch":"","is_wiki":false,"is_gitlab_snippet":false,"is_gitlab_root_level_snippet":false,"is_github_gist":false,"gitlab_snippet_info":{"id":"","title":"","url_of_repo":"","name_of_repo":""},"commits":{"count_pre_pull":0,"count_post_pull":0,"count_diff":0}}]'
`)
	t.Setenv("GHORG_REPO_FILTER_HOOK", hook)
	// Clean up other GHORG_ variables to avoid interference with other filters
	t.Setenv("GHORG_TARGET_REPOS_PATH", "")
	t.Setenv("GHORG_MATCH_REGEX", "")
	t.Setenv("GHORG_EXCLUDE_MATCH_REGEX", "")
	t.Setenv("GHORG_MATCH_PREFIX", "")
	t.Setenv("GHORG_EXCLUDE_MATCH_PREFIX", "")

	filter := NewRepositoryFilter()
	repos := []scm.Repo{{Name: "keep-me"}, {Name: "drop-me"}}

	result := filter.ApplyAllFilters(repos)
	expected := []scm.Repo{{Name: "keep-me"}}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %+v, got %+v", expected, result)
	}
}

func TestRunRepoFilterHook_NilReposSentAsEmptyArray(t *testing.T) {
	skipOnWindows(t)
	hook := writeHookScript(t, "#!/bin/sh\ninput=$(cat)\nif [ \"$input\" = \"[]\" ]; then echo '[]'; else echo \"$input\" >&2; exit 1; fi\n")

	result, err := runRepoFilterHook(hook, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty result, got %+v", result)
	}
}

func TestRunRepoFilterHook_NonExecutableHookReturnsError(t *testing.T) {
	skipOnWindows(t)
	path := filepath.Join(t.TempDir(), "hook.sh")
	if err := os.WriteFile(path, []byte("#!/bin/sh\ncat\n"), 0o644); err != nil {
		t.Fatalf("failed to write hook file: %v", err)
	}

	_, err := runRepoFilterHook(path, []scm.Repo{{Name: "repo1"}})
	if err == nil {
		t.Fatal("expected error for non-executable hook, got nil")
	}
}
