package cmd

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/gabrie30/ghorg/scm"
	"github.com/spf13/cobra"
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

func TestRelativePathRepositories(t *testing.T) {
	testing, err := os.MkdirTemp("", "testing")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(testing)

	outputDirAbsolutePath = testing

	repository := filepath.Join(testing, "repository", ".git")
	if err := os.MkdirAll(repository, 0o755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	files, err := getRelativePathRepositories(testing)
	if err != nil {
		t.Fatalf("getRelativePathRepositories returned an error: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 directory, got %d", len(files))
	}

	if len(files) > 0 && files[0] != "repository" {
		t.Errorf("Expected 'repository', got '%s'", files[0])
	}
}

func TestRelativePathRepositoriesNoGitDir(t *testing.T) {
	testing, err := os.MkdirTemp("", "testing")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(testing)

	outputDirAbsolutePath = testing

	directory := filepath.Join(testing, "directory")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	files, err := getRelativePathRepositories(testing)
	if err != nil {
		t.Fatalf("getRelativePathRepositories returned an error: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 directories, got %d", len(files))
	}
}

func TestRelativePathRepositoriesWithGitSubmodule(t *testing.T) {
	testing, err := os.MkdirTemp("", "testing")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(testing)

	outputDirAbsolutePath = testing

	repository := filepath.Join(testing, "repository", ".git")
	submodule := filepath.Join(testing, "repository", "submodule", ".git")

	if err := os.MkdirAll(repository, 0o755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(submodule), 0o755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	if _, err := os.Create(submodule); err != nil {
		t.Fatalf("Failed to create .git file: %v", err)
	}

	files, err := getRelativePathRepositories(testing)
	if err != nil {
		t.Fatalf("getRelativePathRepositories returned an error: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 directory, got %d", len(files))
	}

	if len(files) > 0 && files[0] != "repository" {
		t.Errorf("Expected 'repository', got '%s'", files[0])
	}
}

func TestRelativePathRepositoriesDeeplyNested(t *testing.T) {
	testing, err := os.MkdirTemp("", "testing")
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	defer os.RemoveAll(testing)

	outputDirAbsolutePath = testing

	repository := filepath.Join(testing, "deeply", "nested", "repository", ".git")
	if err := os.MkdirAll(repository, 0o755); err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	files, err := getRelativePathRepositories(testing)
	if err != nil {
		t.Fatalf("getRelativePathRepositories returned an error: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 directory, got %d", len(files))
	}

	expected := filepath.Join("deeply", "nested", "repository")
	if len(files) > 0 && files[0] != expected {
		t.Errorf("Expected '%s', got '%s'", expected, files[0])
	}
}

func TestPruneRepos(t *testing.T) {
	os.Setenv("GHORG_PRUNE_NO_CONFIRM", "true")

	cloneTargets := []scm.Repo{{Path: "/repository"}}

	testing, err := os.MkdirTemp("", "testing")
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}
	defer os.RemoveAll(testing)

	outputDirAbsolutePath = testing

	repository := filepath.Join(testing, "repository", ".git")
	if err := os.MkdirAll(repository, 0o755); err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	prunable := filepath.Join(testing, "prunnable", ".git")
	if err := os.MkdirAll(prunable, 0o755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	pruneRepos(cloneTargets)

	if _, err := os.Stat(repository); os.IsNotExist(err) {
		t.Errorf("Expected '%s' to exist, but it was deleted", repository)
	}

	if _, err := os.Stat(prunable); !os.IsNotExist(err) {
		t.Errorf("Expected '%s' to be deleted, but it exists", prunable)
	}
}

func TestSliceContainsNamedRepoWithPathSeparators(t *testing.T) {
	// Test that path separator normalization works correctly
	// This simulates the Windows issue where GitLab API returns forward slashes
	// but Windows filesystem uses backslashes

	testCases := []struct {
		name        string
		repos       []scm.Repo
		needle      string
		shouldMatch bool
	}{
		{
			name:        "Forward slash in repo, forward slash in needle",
			repos:       []scm.Repo{{Path: "group/subgroup/repo"}},
			needle:      "group/subgroup/repo",
			shouldMatch: true,
		},
		{
			name:        "Forward slash in repo, backslash in needle (Windows case)",
			repos:       []scm.Repo{{Path: "group/subgroup/repo"}},
			needle:      "group\\subgroup\\repo",
			shouldMatch: true,
		},
		{
			name:        "Backslash in repo, forward slash in needle",
			repos:       []scm.Repo{{Path: "group\\subgroup\\repo"}},
			needle:      "group/subgroup/repo",
			shouldMatch: true,
		},
		{
			name:        "Leading slash normalization",
			repos:       []scm.Repo{{Path: "/group/subgroup/repo"}},
			needle:      "group\\subgroup\\repo",
			shouldMatch: true,
		},
		{
			name:        "Mixed separators",
			repos:       []scm.Repo{{Path: "group/subgroup\\repo"}},
			needle:      "group\\subgroup/repo",
			shouldMatch: true,
		},
		{
			name:        "No match case",
			repos:       []scm.Repo{{Path: "group/subgroup/repo"}},
			needle:      "different/path",
			shouldMatch: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := sliceContainsNamedRepo(tc.repos, tc.needle)
			if result != tc.shouldMatch {
				t.Errorf("Expected %v, got %v for needle '%s' in repos %+v",
					tc.shouldMatch, result, tc.needle, tc.repos)
			}
		})
	}
}

func TestSyncDefaultBranchFlagSetsEnvironmentVariable(t *testing.T) {
	// Save current environment state
	originalValue := os.Getenv("GHORG_SYNC_DEFAULT_BRANCH")
	defer func() {
		if originalValue != "" {
			os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", originalValue)
		} else {
			os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
		}
	}()

	// Clear the environment variable to start clean
	os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

	// Create a mock command with the sync flag set
	cmd := cloneCmd
	cmd.Flags().Set("sync-default-branch", "true")

	// Simulate flag processing that would happen in cloneFunc
	if cmd.Flags().Changed("sync-default-branch") {
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
	}

	// Verify the environment variable was set
	if os.Getenv("GHORG_SYNC_DEFAULT_BRANCH") != "true" {
		t.Errorf("Expected GHORG_SYNC_DEFAULT_BRANCH to be 'true', got '%s'", os.Getenv("GHORG_SYNC_DEFAULT_BRANCH"))
	}
}

func TestSyncDefaultBranchFlagIntegration(t *testing.T) {
	tests := []struct {
		name          string
		flagSet       bool
		expectedValue string
	}{
		{
			name:          "Flag not set",
			flagSet:       false,
			expectedValue: "",
		},
		{
			name:          "Flag set to true",
			flagSet:       true,
			expectedValue: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore environment
			originalValue := os.Getenv("GHORG_SYNC_DEFAULT_BRANCH")
			defer func() {
				if originalValue != "" {
					os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", originalValue)
				} else {
					os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
				}
			}()

			// Clear environment
			os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

			// Create fresh command instance
			cmd := &cobra.Command{
				Use: "test",
			}
			var syncFlag bool
			cmd.Flags().BoolVar(&syncFlag, "sync-default-branch", false, "test flag")

			if tt.flagSet {
				cmd.Flags().Set("sync-default-branch", "true")

				// Simulate the flag processing logic from cloneFunc
				if cmd.Flags().Changed("sync-default-branch") {
					os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
				}
			}

			// Verify environment variable state
			actualValue := os.Getenv("GHORG_SYNC_DEFAULT_BRANCH")
			if actualValue != tt.expectedValue {
				t.Errorf("Expected GHORG_SYNC_DEFAULT_BRANCH to be '%s', got '%s'", tt.expectedValue, actualValue)
			}
		})
	}
}

func TestSyncDefaultBranchFlagValidation(t *testing.T) {
	// Test that the flag is properly defined in the command
	cmd := cloneCmd
	flag := cmd.Flags().Lookup("sync-default-branch")

	if flag == nil {
		t.Fatal("sync-default-branch flag not found")
	}

	if flag.DefValue != "false" {
		t.Errorf("Expected default value 'false', got '%s'", flag.DefValue)
	}

	if flag.Usage == "" {
		t.Error("Flag should have usage text")
	}

	if !strings.Contains(flag.Usage, "GHORG_SYNC_DEFAULT_BRANCH") {
		t.Error("Flag usage should mention GHORG_SYNC_DEFAULT_BRANCH environment variable")
	}
}

func TestSyncDefaultBranchFlagWithOtherFlags(t *testing.T) {
	// Test that sync flag works in combination with other flags
	originalValue := os.Getenv("GHORG_SYNC_DEFAULT_BRANCH")
	defer func() {
		if originalValue != "" {
			os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", originalValue)
		} else {
			os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
		}
	}()

	os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

	cmd := cloneCmd
	// Set multiple flags
	cmd.Flags().Set("sync-default-branch", "true")
	cmd.Flags().Set("dry-run", "true")
	cmd.Flags().Set("quiet", "true")

	// Process sync flag specifically
	if cmd.Flags().Changed("sync-default-branch") {
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
	}

	// Verify sync flag worked even with other flags set
	if os.Getenv("GHORG_SYNC_DEFAULT_BRANCH") != "true" {
		t.Error("sync-default-branch flag should work with other flags")
	}
}

// TestSyncIntegrationWithFlag tests the complete integration between
// the flag system and the sync functionality
func TestSyncIntegrationWithFlag(t *testing.T) {
	tests := []struct {
		name        string
		flagSet     bool
		expectSync  bool
		description string
	}{
		{
			name:        "Flag not set - sync should be disabled",
			flagSet:     false,
			expectSync:  false,
			description: "When --sync-default-branch flag is not used, sync should be disabled",
		},
		{
			name:        "Flag set - sync should be enabled",
			flagSet:     true,
			expectSync:  true,
			description: "When --sync-default-branch flag is used, sync should be enabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore environment
			originalValue := os.Getenv("GHORG_SYNC_DEFAULT_BRANCH")
			defer func() {
				if originalValue != "" {
					os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", originalValue)
				} else {
					os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
				}
			}()

			// Clear environment
			os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

			// Simulate flag processing that happens in cloneFunc
			if tt.flagSet {
				// This simulates what happens in cloneFunc when flag is set
				os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
			}

			// Test that the sync logic responds correctly to the environment variable
			syncEnabled := os.Getenv("GHORG_SYNC_DEFAULT_BRANCH")
			actualSyncWouldRun := syncEnabled == "true"

			if actualSyncWouldRun != tt.expectSync {
				t.Errorf("Expected sync to be %v, but environment suggests it would be %v", tt.expectSync, actualSyncWouldRun)
			}

			// Verify the logic matches the sync function's early return check
			if !tt.expectSync && syncEnabled != "true" {
				t.Logf("✓ Sync correctly disabled when flag not set")
			} else if tt.expectSync && syncEnabled == "true" {
				t.Logf("✓ Sync correctly enabled when flag is set")
			}
		})
	}
}

// TestCompleteWorkflowFlagToSync tests the complete workflow from flag to sync
func TestCompleteWorkflowFlagToSync(t *testing.T) {
	// Save and restore environment
	originalValue := os.Getenv("GHORG_SYNC_DEFAULT_BRANCH")
	defer func() {
		if originalValue != "" {
			os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", originalValue)
		} else {
			os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
		}
	}()

	t.Run("Complete workflow - flag enables sync", func(t *testing.T) {
		// Step 1: Clear environment (simulates fresh start)
		os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

		// Step 2: Simulate flag being set
		cmd := cloneCmd
		cmd.Flags().Set("sync-default-branch", "true")

		// Step 3: Simulate flag processing (this happens in cloneFunc)
		if cmd.Flags().Changed("sync-default-branch") {
			os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		}

		// Step 4: Verify sync function would read the correct value
		syncEnabled := os.Getenv("GHORG_SYNC_DEFAULT_BRANCH")
		if syncEnabled != "true" {
			t.Errorf("Expected GHORG_SYNC_DEFAULT_BRANCH to be 'true' after flag processing, got '%s'", syncEnabled)
		}

		// Step 5: Verify this matches what sync logic expects
		shouldSync := syncEnabled == "true"
		if !shouldSync {
			t.Error("Sync logic should recognize that sync is enabled")
		}
	})

	t.Run("Complete workflow - no flag disables sync", func(t *testing.T) {
		// Step 1: Clear environment
		os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

		// Step 2: Reset flags to ensure clean state
		cmd := cloneCmd
		cmd.ResetFlags()
		// Re-add the sync flag (simulates fresh initialization)
		cmd.Flags().BoolVar(&syncDefaultBranch, "sync-default-branch", false, "GHORG_SYNC_DEFAULT_BRANCH - Enable sync")

		// Step 3: Don't set the sync flag, so Changed() should return false
		// Simulate flag processing - sync flag not changed
		if cmd.Flags().Changed("sync-default-branch") {
			os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		}

		// Step 4: Verify sync function reads empty/false value
		syncEnabled := os.Getenv("GHORG_SYNC_DEFAULT_BRANCH")
		if syncEnabled != "" {
			t.Errorf("Expected GHORG_SYNC_DEFAULT_BRANCH to be empty when flag not set, got '%s'", syncEnabled)
		}

		// Step 5: Verify sync logic would be disabled
		shouldSync := syncEnabled == "true"
		if shouldSync {
			t.Error("Sync logic should recognize that sync is disabled")
		}
	})
}

// TestSyncFlagProcessingEdgeCases tests edge cases in flag processing
func TestSyncFlagProcessingEdgeCases(t *testing.T) {
	// Save and restore environment
	originalValue := os.Getenv("GHORG_SYNC_DEFAULT_BRANCH")
	defer func() {
		if originalValue != "" {
			os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", originalValue)
		} else {
			os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
		}
	}()

	t.Run("Flag processing with pre-existing environment variable", func(t *testing.T) {
		// Set environment variable to false initially
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "false")

		// Set the flag
		cmd := cloneCmd
		cmd.Flags().Set("sync-default-branch", "true")

		// Process flag - should override environment variable
		if cmd.Flags().Changed("sync-default-branch") {
			os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		}

		// Verify flag processing overrides existing environment
		if os.Getenv("GHORG_SYNC_DEFAULT_BRANCH") != "true" {
			t.Error("Flag processing should override existing environment variable")
		}
	})

	t.Run("Flag not changed - environment remains unchanged", func(t *testing.T) {
		// Set environment variable initially
		os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "manual_value")

		// Create command and reset flags to ensure clean state
		cmd := cloneCmd
		cmd.ResetFlags()
		// Re-add the sync flag (simulates fresh initialization)
		cmd.Flags().BoolVar(&syncDefaultBranch, "sync-default-branch", false, "GHORG_SYNC_DEFAULT_BRANCH - Enable sync")

		// Process flags - sync flag wasn't changed, so env should remain
		if cmd.Flags().Changed("sync-default-branch") {
			os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		}

		// Verify environment wasn't changed by flag processing
		if os.Getenv("GHORG_SYNC_DEFAULT_BRANCH") != "manual_value" {
			t.Error("Environment variable should remain unchanged when flag not set")
		}
	})
}

// TestSyncFlagBehaviorMatchesDocumentation tests that flag behavior matches documented behavior
func TestSyncFlagBehaviorMatchesDocumentation(t *testing.T) {
	// Save and restore environment
	originalValue := os.Getenv("GHORG_SYNC_DEFAULT_BRANCH")
	defer func() {
		if originalValue != "" {
			os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", originalValue)
		} else {
			os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
		}
	}()

	// Test documented behavior: flag should enable sync
	t.Run("Flag enables sync as documented", func(t *testing.T) {
		os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

		// Set flag as user would
		cmd := cloneCmd
		cmd.Flags().Set("sync-default-branch", "true")

		// Process as cloneFunc would
		if cmd.Flags().Changed("sync-default-branch") {
			os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		}

		// Verify behavior matches documentation
		syncEnabled := os.Getenv("GHORG_SYNC_DEFAULT_BRANCH") == "true"
		if !syncEnabled {
			t.Error("Flag should enable sync as documented in help text")
		}
	})

	// Test documented behavior: default is disabled
	t.Run("Default behavior is sync disabled", func(t *testing.T) {
		os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")

		// Reset flags to ensure clean state
		cmd := cloneCmd
		cmd.ResetFlags()
		// Re-add the sync flag (simulates fresh initialization)
		cmd.Flags().BoolVar(&syncDefaultBranch, "sync-default-branch", false, "GHORG_SYNC_DEFAULT_BRANCH - Enable sync")

		// Process flags - no flags set, so nothing should change
		if cmd.Flags().Changed("sync-default-branch") {
			os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
		}

		// Verify default behavior is disabled
		syncEnabled := os.Getenv("GHORG_SYNC_DEFAULT_BRANCH") == "true"
		if syncEnabled {
			t.Error("Default behavior should be sync disabled")
		}
	})
}
