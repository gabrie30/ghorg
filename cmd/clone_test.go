package cmd

import (
	"errors"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/gabrie30/ghorg/scm"
)

func TestShouldLowerRegularString(t *testing.T) {

	upperName := "RepoName"
	defer setOutputDirName([]string{""})
	setOutputDirName([]string{upperName})

	if outputDirName != "reponame" {
		t.Errorf("Wrong folder name, expected: %s, got: %s", upperName, outputDirName)
	}
}

func TestShouldNotChangeLowerCasedRegularString(t *testing.T) {

	lowerName := "repo_name"
	defer setOutputDirName([]string{""})
	setOutputDirName([]string{lowerName})

	if outputDirName != "repo_name" {
		t.Errorf("Wrong folder name, expected: %s, got: %s", lowerName, outputDirName)
	}
}

func TestReplaceDashWithUnderscore(t *testing.T) {

	want := "repo-name"
	lowerName := "repo-name"
	defer setOutputDirName([]string{""})
	setOutputDirName([]string{lowerName})

	if outputDirName != want {
		t.Errorf("Wrong folder name, expected: %s, got: %s", want, outputDirName)
	}
}

func TestShouldNotChangeNonLettersString(t *testing.T) {

	numberName := "1234567_8"
	defer setOutputDirName([]string{""})
	setOutputDirName([]string{numberName})

	if outputDirName != "1234567_8" {
		t.Errorf("Wrong folder name, expected: %s, got: %s", numberName, outputDirName)
	}
}

type MockGitClient struct{}

func NewMockGit() MockGitClient {
	return MockGitClient{}
}

func (g MockGitClient) HasRemoteHeads(repo scm.Repo) (bool, error) {
	if repo.Name == "testRepoEmpty" {
		return false, nil
	}
	return true, nil
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
	if repo.Name == "testRepoEmpty" {
		return errors.New("Cannot checkout any specific branch in an empty repository")
	}
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

func (g MockGitClient) FetchCloneBranch(repo scm.Repo) error {
	return nil
}

func (g MockGitClient) RepoCommitCount(repo scm.Repo) (int, error) {
	return 0, nil
}

func (g MockGitClient) Branch(repo scm.Repo) (string, error) {
	return "", nil
}

func (g MockGitClient) RevListCompare(repo scm.Repo, ref1 string, ref2 string) (string, error) {
	return "", nil
}

func (g MockGitClient) ShortStatus(repo scm.Repo) (string, error) {
	return "", nil
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

func TestCloneEmptyRepo(t *testing.T) {
	defer UnsetEnv("GHORG_")()
	dir, err := os.MkdirTemp("", "ghorg_test_empty_repo")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	os.Setenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO", dir)
	setOuputDirAbsolutePath()

	os.Setenv("GHORG_DONT_EXIT_UNDER_TEST", "true")

	// simulate a previous clone of empty git repository
	repoErr := os.Mkdir(outputDirAbsolutePath+"/"+"testRepoEmpty", 0o700)
	if repoErr != nil {
		log.Fatal(repoErr)
	}
	defer os.RemoveAll(outputDirAbsolutePath + "/" + "testRepoEmpty")

	os.Setenv("GHORG_CONCURRENCY", "1")
	var testRepos = []scm.Repo{
		{
			Name:        "testRepoEmpty",
			URL:         "git@github.com:org/testRepoEmpty.git",
			CloneBranch: "main",
		},
	}

	mockGit := NewMockGit()
	CloneAllRepos(mockGit, testRepos)
	gotInfos := len(cloneInfos)
	expectedInfos := 1
	if gotInfos != expectedInfos {
		t.Fatalf("Wrong number of cloneInfos, expected: %v, got: %v", expectedInfos, gotInfos)
	}
	gotInfo := cloneInfos[0]
	expected := "Could not checkout main due to repository being empty, no changes made on: git@github.com:org/testRepoEmpty.git"
	if gotInfo != expected {
		t.Errorf("Wrong cloneInfo, expected: %v, got: %v", expected, gotInfo)
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
	os.Setenv("GHORG_DONT_EXIT_UNDER_TEST", "true")

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
	os.Setenv("GHORG_DONT_EXIT_UNDER_TEST", "true")

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
	os.Setenv("GHORG_DONT_EXIT_UNDER_TEST", "true")

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
	os.Setenv("GHORG_DONT_EXIT_UNDER_TEST", "true")

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

func Test_filterWithGhorgignore(t *testing.T) {
	type testCase struct {
		name           string
		cloneTargets   []scm.Repo
		expectedResult []scm.Repo
	}

	testCases := []testCase{
		{
			name: "filters out repo named 'shouldbeignored'",
			cloneTargets: []scm.Repo{
				{Name: "shouldbeignored", URL: "https://github.com/org/shouldbeignored"},
				{Name: "bar", URL: "https://github.com/org/bar"},
			},
			expectedResult: []scm.Repo{
				{Name: "bar", URL: "https://github.com/org/bar"},
			},
		},
		{
			name: "filters out repo named 'shouldbeignored'",
			cloneTargets: []scm.Repo{
				{Name: "foo", URL: "https://github.com/org/foo"},
				{Name: "shouldbeignored", URL: "https://github.com/org/shouldbeignored"},
			},
			expectedResult: []scm.Repo{
				{Name: "foo", URL: "https://github.com/org/foo"},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := createTempFileWithContent("shouldbeignored")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())

			os.Setenv("GHORG_IGNORE_PATH", tmpfile.Name())

			got := filterByGhorgignore(tt.cloneTargets)
			if !reflect.DeepEqual(got, tt.expectedResult) {
				t.Errorf("filterWithGhorgignore() = %v, want %v", got, tt.expectedResult)
			}
		})
	}
}

// createTempFileWithContent will create
func createTempFileWithContent(content string) (*os.File, error) {
	tmpfile, err := os.CreateTemp("", "ghorgtest")
	if err != nil {
		return nil, err
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		return nil, err
	}

	if err := tmpfile.Close(); err != nil {
		return nil, err
	}

	return tmpfile, nil
}

func Test_filterDownReposIfTargetReposPathEnabled(t *testing.T) {
	type testCase struct {
		name           string
		cloneTargets   []scm.Repo
		expectedResult []scm.Repo
	}

	testCases := []testCase{
		{
			name: "filters out repos not matching 'targetRepo'",
			cloneTargets: []scm.Repo{
				{Name: "targetRepo", URL: "https://github.com/org/targetRepo"},
				{Name: "bar", URL: "https://github.com/org/bar"},
			},
			expectedResult: []scm.Repo{
				{Name: "targetRepo", URL: "https://github.com/org/targetRepo"},
			},
		},
		{
			name: "filters out all repos",
			cloneTargets: []scm.Repo{
				{Name: "foo", URL: "https://github.com/org/foo"},
				{Name: "shouldbefiltered", URL: "https://github.com/org/shouldbefiltered"},
			},
			expectedResult: []scm.Repo{},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := createTempFileWithContent("targetRepo")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())

			os.Setenv("GHORG_TARGET_REPOS_PATH", tmpfile.Name())

			got := filterByTargetReposPath(tt.cloneTargets)
			if !reflect.DeepEqual(got, tt.expectedResult) {
				t.Errorf("filterWithGhorgignore() = %v, want %v", got, tt.expectedResult)
			}
		})
	}
}
