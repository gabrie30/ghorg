package cmd

import (
	"os"
	"reflect"
	"testing"

	"github.com/gabrie30/ghorg/scm"
)

func TestRepositoryFilter_FilterByRegexMatch(t *testing.T) {
	filter := NewRepositoryFilter()

	testCases := []struct {
		name          string
		regex         string
		repos         []scm.Repo
		expectedRepos []scm.Repo
	}{
		{
			name:  "matches repos with prefix",
			regex: "^test-",
			repos: []scm.Repo{
				{Name: "test-repo1"},
				{Name: "test-repo2"},
				{Name: "other-repo"},
			},
			expectedRepos: []scm.Repo{
				{Name: "test-repo1"},
				{Name: "test-repo2"},
			},
		},
		{
			name:  "matches repos with suffix",
			regex: "-lib$",
			repos: []scm.Repo{
				{Name: "utils-lib"},
				{Name: "core-lib"},
				{Name: "main-app"},
			},
			expectedRepos: []scm.Repo{
				{Name: "utils-lib"},
				{Name: "core-lib"},
			},
		},
		{
			name:          "no matches",
			regex:         "^nonexistent",
			repos:         []scm.Repo{{Name: "repo1"}, {Name: "repo2"}},
			expectedRepos: []scm.Repo{},
		},
		{
			name:          "empty regex returns all",
			regex:         "",
			repos:         []scm.Repo{{Name: "repo1"}, {Name: "repo2"}},
			expectedRepos: []scm.Repo{{Name: "repo1"}, {Name: "repo2"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("GHORG_MATCH_REGEX", tc.regex)
			defer os.Unsetenv("GHORG_MATCH_REGEX")

			result := filter.FilterByRegexMatch(tc.repos)
			if !reflect.DeepEqual(result, tc.expectedRepos) {
				t.Errorf("Expected %v, got %v", tc.expectedRepos, result)
			}
		})
	}
}

func TestRepositoryFilter_FilterByExcludeRegexMatch(t *testing.T) {
	filter := NewRepositoryFilter()

	testCases := []struct {
		name          string
		regex         string
		repos         []scm.Repo
		expectedRepos []scm.Repo
	}{
		{
			name:  "excludes repos with prefix",
			regex: "^test-",
			repos: []scm.Repo{
				{Name: "test-repo1"},
				{Name: "test-repo2"},
				{Name: "other-repo"},
			},
			expectedRepos: []scm.Repo{
				{Name: "other-repo"},
			},
		},
		{
			name:  "excludes repos with suffix",
			regex: "-test$",
			repos: []scm.Repo{
				{Name: "utils-test"},
				{Name: "core-lib"},
				{Name: "main-test"},
			},
			expectedRepos: []scm.Repo{
				{Name: "core-lib"},
			},
		},
		{
			name:          "no exclusions",
			regex:         "^nonexistent",
			repos:         []scm.Repo{{Name: "repo1"}, {Name: "repo2"}},
			expectedRepos: []scm.Repo{{Name: "repo1"}, {Name: "repo2"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("GHORG_EXCLUDE_MATCH_REGEX", tc.regex)
			defer os.Unsetenv("GHORG_EXCLUDE_MATCH_REGEX")

			result := filter.FilterByExcludeRegexMatch(tc.repos)
			if !reflect.DeepEqual(result, tc.expectedRepos) {
				t.Errorf("Expected %v, got %v", tc.expectedRepos, result)
			}
		})
	}
}

func TestRepositoryFilter_FilterByMatchPrefix(t *testing.T) {
	filter := NewRepositoryFilter()

	testCases := []struct {
		name          string
		prefix        string
		repos         []scm.Repo
		expectedRepos []scm.Repo
	}{
		{
			name:   "matches single prefix",
			prefix: "test",
			repos: []scm.Repo{
				{Name: "test-repo1"},
				{Name: "Test-Repo2"}, // Should match case-insensitive
				{Name: "other-repo"},
			},
			expectedRepos: []scm.Repo{
				{Name: "test-repo1"},
				{Name: "Test-Repo2"},
			},
		},
		{
			name:   "matches multiple prefixes",
			prefix: "test,lib",
			repos: []scm.Repo{
				{Name: "test-repo1"},
				{Name: "lib-utils"},
				{Name: "other-repo"},
				{Name: "lib-core"},
			},
			expectedRepos: []scm.Repo{
				{Name: "test-repo1"},
				{Name: "lib-utils"},
				{Name: "lib-core"},
			},
		},
		{
			name:          "no matches",
			prefix:        "nonexistent",
			repos:         []scm.Repo{{Name: "repo1"}, {Name: "repo2"}},
			expectedRepos: []scm.Repo{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("GHORG_MATCH_PREFIX", tc.prefix)
			defer os.Unsetenv("GHORG_MATCH_PREFIX")

			result := filter.FilterByMatchPrefix(tc.repos)
			if !reflect.DeepEqual(result, tc.expectedRepos) {
				t.Errorf("Expected %v, got %v", tc.expectedRepos, result)
			}
		})
	}
}

func TestRepositoryFilter_FilterByExcludeMatchPrefix(t *testing.T) {
	filter := NewRepositoryFilter()

	testCases := []struct {
		name          string
		prefix        string
		repos         []scm.Repo
		expectedRepos []scm.Repo
	}{
		{
			name:   "excludes single prefix",
			prefix: "test",
			repos: []scm.Repo{
				{Name: "test-repo1"},
				{Name: "Test-Repo2"}, // Should exclude case-insensitive
				{Name: "other-repo"},
			},
			expectedRepos: []scm.Repo{
				{Name: "other-repo"},
			},
		},
		{
			name:   "excludes multiple prefixes",
			prefix: "test,lib",
			repos: []scm.Repo{
				{Name: "test-repo1"},
				{Name: "lib-utils"},
				{Name: "other-repo"},
				{Name: "main-app"},
			},
			expectedRepos: []scm.Repo{
				{Name: "other-repo"},
				{Name: "main-app"},
			},
		},
		{
			name:          "no exclusions",
			prefix:        "nonexistent",
			repos:         []scm.Repo{{Name: "repo1"}, {Name: "repo2"}},
			expectedRepos: []scm.Repo{{Name: "repo1"}, {Name: "repo2"}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("GHORG_EXCLUDE_MATCH_PREFIX", tc.prefix)
			defer os.Unsetenv("GHORG_EXCLUDE_MATCH_PREFIX")

			result := filter.FilterByExcludeMatchPrefix(tc.repos)
			if !reflect.DeepEqual(result, tc.expectedRepos) {
				t.Errorf("Expected %v, got %v", tc.expectedRepos, result)
			}
		})
	}
}

func TestRepositoryFilter_FilterByGhorgignore(t *testing.T) {
	filter := NewRepositoryFilter()

	testCases := []struct {
		name          string
		ignoreContent string
		repos         []scm.Repo
		expectedRepos []scm.Repo
	}{
		{
			name:          "filters out matching URLs",
			ignoreContent: "shouldbeignored",
			repos: []scm.Repo{
				{Name: "repo1", URL: "https://github.com/org/repo1"},
				{Name: "shouldbeignored", URL: "https://github.com/org/shouldbeignored"},
			},
			expectedRepos: []scm.Repo{
				{Name: "repo1", URL: "https://github.com/org/repo1"},
			},
		},
		{
			name:          "filters multiple patterns",
			ignoreContent: "test-repo\nold-project",
			repos: []scm.Repo{
				{Name: "repo1", URL: "https://github.com/org/repo1"},
				{Name: "test-repo", URL: "https://github.com/org/test-repo"},
				{Name: "old-project", URL: "https://github.com/org/old-project"},
			},
			expectedRepos: []scm.Repo{
				{Name: "repo1", URL: "https://github.com/org/repo1"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary ignore file
			tmpfile, err := createTempFileWithContent(tc.ignoreContent)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())

			os.Setenv("GHORG_IGNORE_PATH", tmpfile.Name())
			defer os.Unsetenv("GHORG_IGNORE_PATH")

			result := filter.FilterByGhorgignore(tc.repos)
			if !reflect.DeepEqual(result, tc.expectedRepos) {
				t.Errorf("Expected %v, got %v", tc.expectedRepos, result)
			}
		})
	}
}

func TestRepositoryFilter_FilterByTargetReposPath(t *testing.T) {
	filter := NewRepositoryFilter()

	testCases := []struct {
		name          string
		targetContent string
		repos         []scm.Repo
		expectedRepos []scm.Repo
	}{
		{
			name:          "filters to target repos only",
			targetContent: "target-repo\nother-target",
			repos: []scm.Repo{
				{Name: "target-repo", URL: "https://github.com/org/target-repo.git"},
				{Name: "other-target", URL: "https://github.com/org/other-target.git"},
				{Name: "unwanted", URL: "https://github.com/org/unwanted.git"},
			},
			expectedRepos: []scm.Repo{
				{Name: "target-repo", URL: "https://github.com/org/target-repo.git"},
				{Name: "other-target", URL: "https://github.com/org/other-target.git"},
			},
		},
		{
			name:          "handles case insensitive matching",
			targetContent: "Target-Repo",
			repos: []scm.Repo{
				{Name: "target-repo", URL: "https://github.com/org/target-repo.git"},
				{Name: "other", URL: "https://github.com/org/other.git"},
			},
			expectedRepos: []scm.Repo{
				{Name: "target-repo", URL: "https://github.com/org/target-repo.git"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary target file
			tmpfile, err := createTempFileWithContent(tc.targetContent)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())

			os.Setenv("GHORG_TARGET_REPOS_PATH", tmpfile.Name())
			defer os.Unsetenv("GHORG_TARGET_REPOS_PATH")

			result := filter.FilterByTargetReposPath(tc.repos)
			if !reflect.DeepEqual(result, tc.expectedRepos) {
				t.Errorf("Expected %v, got %v", tc.expectedRepos, result)
			}
		})
	}
}

func TestRepositoryFilter_ApplyAllFilters(t *testing.T) {
	defer UnsetEnv("GHORG_")()

	filter := NewRepositoryFilter()
	repos := []scm.Repo{
		{Name: "test-repo1", URL: "https://github.com/org/test-repo1.git"},
		{Name: "test-repo2", URL: "https://github.com/org/test-repo2.git"},
		{Name: "lib-utils", URL: "https://github.com/org/lib-utils.git"},
		{Name: "ignored", URL: "https://github.com/org/ignored.git"},
		{Name: "other", URL: "https://github.com/org/other.git"},
	}

	// Set up regex filter to match test- prefix
	os.Setenv("GHORG_MATCH_REGEX", "^test-")

	// Set up ghorgignore
	tmpfile, err := createTempFileWithContent("ignored")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	os.Setenv("GHORG_IGNORE_PATH", tmpfile.Name())

	result := filter.ApplyAllFilters(repos)

	expected := []scm.Repo{
		{Name: "test-repo1", URL: "https://github.com/org/test-repo1.git"},
		{Name: "test-repo2", URL: "https://github.com/org/test-repo2.git"},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// Benchmark tests for performance validation
func BenchmarkRepositoryFilter_FilterByRegexMatch(b *testing.B) {
	filter := NewRepositoryFilter()
	os.Setenv("GHORG_MATCH_REGEX", "^test-")
	defer os.Unsetenv("GHORG_MATCH_REGEX")

	// Create 1000 test repos
	repos := make([]scm.Repo, 1000)
	for i := 0; i < 1000; i++ {
		if i%2 == 0 {
			repos[i] = scm.Repo{Name: "test-repo" + string(rune(i))}
		} else {
			repos[i] = scm.Repo{Name: "other-repo" + string(rune(i))}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.FilterByRegexMatch(repos)
	}
}

func BenchmarkRepositoryFilter_FilterByPrefix(b *testing.B) {
	filter := NewRepositoryFilter()
	os.Setenv("GHORG_MATCH_PREFIX", "test,lib,core")
	defer os.Unsetenv("GHORG_MATCH_PREFIX")

	// Create 1000 test repos
	repos := make([]scm.Repo, 1000)
	prefixes := []string{"test", "lib", "core", "other", "main"}
	for i := 0; i < 1000; i++ {
		repos[i] = scm.Repo{Name: prefixes[i%5] + "-repo" + string(rune(i))}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.FilterByMatchPrefix(repos)
	}
}
