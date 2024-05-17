package cmd

import (
	"log"
	"os"
	"strings"
	"testing"

	"github.com/gabrie30/ghorg/scm"
)

func TestShouldLowerRegularString(t *testing.T) {

	upperName := "RepoName"
	setOutputDirName([]string{upperName})

	if outputDirName != "reponame" {
		t.Errorf("Wrong folder name, expected: %s, got: %s", upperName, outputDirName)
	}
}

func TestShouldNotChangeLowerCasedRegularString(t *testing.T) {

	lowerName := "repo_name"
	setOutputDirName([]string{lowerName})

	if outputDirName != "repo_name" {
		t.Errorf("Wrong folder name, expected: %s, got: %s", lowerName, outputDirName)
	}
}

func TestReplaceDashWithUnderscore(t *testing.T) {

	want := "repo-name"
	lowerName := "repo-name"
	setOutputDirName([]string{lowerName})

	if outputDirName != want {
		t.Errorf("Wrong folder name, expected: %s, got: %s", want, outputDirName)
	}
}

func TestShouldNotChangeNonLettersString(t *testing.T) {

	numberName := "1234567_8"
	setOutputDirName([]string{numberName})

	if outputDirName != "1234567_8" {
		t.Errorf("Wrong folder name, expected: %s, got: %s", numberName, outputDirName)
	}
}

type MockGitClient struct{}

func NewMockGit() MockGitClient {
	return MockGitClient{}
}

func (g MockGitClient) Clone(repo scm.Repo) error {
	_, err := os.MkdirTemp(os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"), repo.Name)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (g MockGitClient) SetOrigin(repo scm.Repo) error {
	return nil
}

func (g MockGitClient) SetOriginWithCredentials(repo scm.Repo) error {
	return nil
}

func (g MockGitClient) Checkout(repo scm.Repo) error {
	return nil
}

func (g MockGitClient) Clean(repo scm.Repo) error {
	return nil
}

func (g MockGitClient) UpdateRemote(repo scm.Repo) error {
	return nil
}

func (g MockGitClient) Pull(repo scm.Repo) error {
	return nil
}

func (g MockGitClient) Reset(repo scm.Repo) error {
	return nil
}

func (g MockGitClient) FetchAll(repo scm.Repo) error {
	return nil
}

func TestInitialClone(t *testing.T) {
	defer UnsetEnv("GHORG_")()
	dir, err := os.MkdirTemp("", "ghorg_test_initial")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO", dir)
	os.Setenv("GHORG_CONCURRENCY", "1")
	var testRepos = []scm.Repo{
		{
			Name: "testRepoOne",
		},
		{
			Name: "testRepoTwo",
		},
	}

	mockGit := NewMockGit()
	CloneAllRepos(mockGit, testRepos)
	got, _ := os.ReadDir(dir)
	expected := len(testRepos)
	if len(got) != expected {
		t.Errorf("Wrong number of repos in clone, expected: %v, got: %v", expected, got)
	}
}

func TestMatchPrefix(t *testing.T) {
	defer UnsetEnv("GHORG_")()
	dir, err := os.MkdirTemp("", "ghorg_test_match_prefix")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO", dir)
	os.Setenv("GHORG_CONCURRENCY", "1")
	os.Setenv("GHORG_MATCH_PREFIX", "test")
	var testRepos = []scm.Repo{
		{
			Name: "testRepoOne",
		},
		{
			Name: "testRepoTwo",
		},
		{
			Name: "testRepoThree",
		},
		{
			Name: "nottestRepoTwo",
		},
		{
			Name: "nottestRepoThree",
		},
	}

	mockGit := NewMockGit()
	CloneAllRepos(mockGit, testRepos)
	got, _ := os.ReadDir(dir)
	expected := 3
	if len(got) != expected {
		t.Errorf("Wrong number of repos in clone, expected: %v, got: %v", expected, len(got))
	}
}

func TestExcludeMatchPrefix(t *testing.T) {
	defer UnsetEnv("GHORG_")()
	dir, err := os.MkdirTemp("", "ghorg_test_exclude_match_prefix")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO", dir)
	os.Setenv("GHORG_CONCURRENCY", "1")
	os.Setenv("GHORG_EXCLUDE_MATCH_PREFIX", "test")
	var testRepos = []scm.Repo{
		{
			Name: "testRepoOne",
		},
		{
			Name: "testRepoTwo",
		},
		{
			Name: "testRepoThree",
		},
		{
			Name: "nottestRepoTwo",
		},
		{
			Name: "nottestRepoThree",
		},
	}

	mockGit := NewMockGit()
	CloneAllRepos(mockGit, testRepos)
	got, _ := os.ReadDir(dir)
	expected := 2
	if len(got) != expected {
		t.Errorf("Wrong number of repos in clone, expected: %v, got: %v", expected, got)
	}
}

func TestMatchRegex(t *testing.T) {
	defer UnsetEnv("GHORG_")()
	dir, err := os.MkdirTemp("", "ghorg_test_match_regex")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO", dir)
	os.Setenv("GHORG_CONCURRENCY", "1")
	os.Setenv("GHORG_MATCH_REGEX", "^test-")
	var testRepos = []scm.Repo{
		{
			Name: "test-RepoOne",
		},
		{
			Name: "test-RepoTwo",
		},
		{
			Name: "test-RepoThree",
		},
		{
			Name: "nottestRepoTwo",
		},
		{
			Name: "nottestRepoThree",
		},
	}

	mockGit := NewMockGit()
	CloneAllRepos(mockGit, testRepos)
	got, _ := os.ReadDir(dir)
	expected := 3
	if len(got) != expected {
		t.Errorf("Wrong number of repos in clone, expected: %v, got: %v", expected, got)
	}
}

func TestExcludeMatchRegex(t *testing.T) {
	defer UnsetEnv("GHORG_")()
	testDescriptor := "ghorg_test_exclude_match_regex"
	dir, err := os.MkdirTemp("", testDescriptor)
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	os.Setenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO", dir)
	os.Setenv("GHORG_CONCURRENCY", "1")
	os.Setenv("GHORG_OUTPUT_DIR", testDescriptor)
	os.Setenv("GHORG_EXCLUDE_MATCH_REGEX", "^test-")
	var testRepos = []scm.Repo{
		{
			Name: "test-RepoOne",
		},
		{
			Name: "test-RepoTwo",
		},
		{
			Name: "test-RepoThree",
		},
		{
			Name: "nottestRepoTwo",
		},
		{
			Name: "nottestRepoThree",
		},
	}

	mockGit := NewMockGit()
	CloneAllRepos(mockGit, testRepos)
	got, _ := os.ReadDir(dir)
	expected := 2
	if len(got) != expected {
		t.Errorf("Wrong number of repos in clone, expected: %v, got: %v", expected, got)
	}
}

// UnsetEnv unsets all envars having prefix and returns a function
// that restores the env. Any newly added envars having prefix are
// also unset by restore. It is idiomatic to use with a defer.
//
// Note that modifying the env may have unpredictable results when
// tests are run with t.Parallel.
func UnsetEnv(prefix string) (restore func()) {
	before := map[string]string{}

	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, prefix) {
			continue
		}

		parts := strings.SplitN(e, "=", 2)
		before[parts[0]] = parts[1]

		os.Unsetenv(parts[0])
	}

	return func() {
		after := map[string]string{}

		for _, e := range os.Environ() {
			if !strings.HasPrefix(e, prefix) {
				continue
			}

			parts := strings.SplitN(e, "=", 2)
			after[parts[0]] = parts[1]

			// Check if the envar previously existed
			v, ok := before[parts[0]]
			if !ok {
				// This is a newly added envar with prefix, zap it
				os.Unsetenv(parts[0])
				continue
			}

			if parts[1] != v {
				// If the envar value has changed, set it back
				os.Setenv(parts[0], v)
			}
		}

		// Still need to check if there have been any deleted envars
		for k, v := range before {
			if _, ok := after[k]; !ok {
				// k is not present in after, so we set it.
				os.Setenv(k, v)
			}
		}
	}
}
