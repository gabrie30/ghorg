package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gabrie30/ghorg/scm"
)

// ExtendedMockGitClient extends the existing MockGitClient with additional methods needed for RepositoryProcessor
type ExtendedMockGitClient struct {
	MockGitClient
	shouldFailClone       bool
	shouldFailCheckout    bool
	shouldFailSetOrigin   bool
	shouldReturnEmptyRepo bool
	preCommitCount        int
	postCommitCount       int
}

func NewExtendedMockGit() *ExtendedMockGitClient {
	return &ExtendedMockGitClient{
		MockGitClient:   NewMockGit(),
		preCommitCount:  5,
		postCommitCount: 7,
	}
}

func (g *ExtendedMockGitClient) Clone(repo scm.Repo) error {
	if g.shouldFailClone {
		return errors.New("mock clone error")
	}
	return g.MockGitClient.Clone(repo)
}

func (g *ExtendedMockGitClient) Checkout(repo scm.Repo) error {
	if g.shouldFailCheckout {
		return errors.New("mock checkout error")
	}
	if g.shouldReturnEmptyRepo {
		return errors.New("Cannot checkout any specific branch in an empty repository")
	}
	return g.MockGitClient.Checkout(repo)
}

func (g *ExtendedMockGitClient) SetOrigin(repo scm.Repo) error {
	if g.shouldFailSetOrigin {
		return errors.New("mock set origin error")
	}
	return g.MockGitClient.SetOrigin(repo)
}

func (g *ExtendedMockGitClient) RepoCommitCount(repo scm.Repo) (int, error) {
	// First call returns pre-pull count, second call returns post-pull count
	if repo.Commits.CountPrePull == 0 {
		return g.preCommitCount, nil
	}
	return g.postCommitCount, nil
}

func TestRepositoryProcessor_NewRepositoryProcessor(t *testing.T) {
	mockGit := NewExtendedMockGit()
	processor := NewRepositoryProcessor(mockGit)

	if processor == nil {
		t.Fatal("Expected processor to be created")
	}

	if processor.git != mockGit {
		t.Error("Expected git client to be set correctly")
	}

	if processor.stats == nil {
		t.Error("Expected stats to be initialized")
	}

	if processor.mutex == nil {
		t.Error("Expected mutex to be initialized")
	}
}

func TestRepositoryProcessor_ProcessRepository_NewRepository(t *testing.T) {
	defer UnsetEnv("GHORG_")()

	// Set up temporary directory
	dir, err := os.MkdirTemp("", "ghorg_test_process_new_repo")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	outputDirAbsolutePath = dir

	mockGit := NewExtendedMockGit()
	processor := NewRepositoryProcessor(mockGit)

	repo := scm.Repo{
		Name:        "test-repo",
		URL:         "https://github.com/org/test-repo",
		CloneBranch: "main",
	}

	repoNameWithCollisions := make(map[string]bool)
	processor.ProcessRepository(&repo, repoNameWithCollisions, false, "test-repo", 0)

	stats := processor.GetStats()
	if stats.CloneCount != 1 {
		t.Errorf("Expected clone count to be 1, got %d", stats.CloneCount)
	}

	if stats.PulledCount != 0 {
		t.Errorf("Expected pulled count to be 0, got %d", stats.PulledCount)
	}

	if len(stats.CloneErrors) != 0 {
		t.Errorf("Expected no clone errors, got %v", stats.CloneErrors)
	}
}

func TestRepositoryProcessor_ProcessRepository_ExistingRepository(t *testing.T) {
	defer UnsetEnv("GHORG_")()

	// Set up temporary directory with existing repo
	dir, err := os.MkdirTemp("", "ghorg_test_process_existing_repo")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	outputDirAbsolutePath = dir

	// Create existing repo directory
	repoDir := filepath.Join(dir, "test-repo")
	err = os.MkdirAll(repoDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	mockGit := NewExtendedMockGit()
	processor := NewRepositoryProcessor(mockGit)

	repo := scm.Repo{
		Name:        "test-repo",
		URL:         "https://github.com/org/test-repo",
		CloneBranch: "main",
		HostPath:    repoDir,
	}

	repoNameWithCollisions := make(map[string]bool)
	processor.ProcessRepository(&repo, repoNameWithCollisions, false, "test-repo", 0)

	stats := processor.GetStats()
	if stats.CloneCount != 0 {
		t.Errorf("Expected clone count to be 0, got %d", stats.CloneCount)
	}

	if stats.PulledCount != 1 {
		t.Errorf("Expected pulled count to be 1, got %d", stats.PulledCount)
	}

	// Check that commit diff was calculated
	if stats.NewCommits != (mockGit.postCommitCount - mockGit.preCommitCount) {
		t.Errorf("Expected new commits to be %d, got %d",
			mockGit.postCommitCount-mockGit.preCommitCount, stats.NewCommits)
	}

	// Verify that CountDiff was properly calculated on the repo
	if repo.Commits.CountDiff != (mockGit.postCommitCount - mockGit.preCommitCount) {
		t.Errorf("Expected repo CountDiff to be %d, got %d",
			mockGit.postCommitCount-mockGit.preCommitCount, repo.Commits.CountDiff)
	}
}

func TestRepositoryProcessor_ProcessRepository_CloneError(t *testing.T) {
	defer UnsetEnv("GHORG_")()

	dir, err := os.MkdirTemp("", "ghorg_test_process_clone_error")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	outputDirAbsolutePath = dir

	mockGit := NewExtendedMockGit()
	mockGit.shouldFailClone = true
	processor := NewRepositoryProcessor(mockGit)

	repo := scm.Repo{
		Name:        "test-repo",
		URL:         "https://github.com/org/test-repo",
		CloneBranch: "main",
	}

	repoNameWithCollisions := make(map[string]bool)
	processor.ProcessRepository(&repo, repoNameWithCollisions, false, "test-repo", 0)

	stats := processor.GetStats()
	if stats.CloneCount != 0 {
		t.Errorf("Expected clone count to be 0, got %d", stats.CloneCount)
	}

	if len(stats.CloneErrors) != 1 {
		t.Errorf("Expected 1 clone error, got %d", len(stats.CloneErrors))
	}

	if stats.CloneErrors[0] == "" {
		t.Error("Expected error message to be set")
	}
}

func TestRepositoryProcessor_ProcessRepository_WikiHandling(t *testing.T) {
	defer UnsetEnv("GHORG_")()

	dir, err := os.MkdirTemp("", "ghorg_test_process_wiki")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	outputDirAbsolutePath = dir

	mockGit := NewExtendedMockGit()
	mockGit.shouldFailClone = true // Simulate wiki with no content
	processor := NewRepositoryProcessor(mockGit)

	repo := scm.Repo{
		Name:        "test-repo",
		URL:         "https://github.com/org/test-repo.wiki",
		CloneBranch: "main",
		IsWiki:      true,
	}

	repoNameWithCollisions := make(map[string]bool)
	processor.ProcessRepository(&repo, repoNameWithCollisions, false, "test-repo.wiki", 0)

	stats := processor.GetStats()
	if len(stats.CloneInfos) != 1 {
		t.Errorf("Expected 1 clone info message, got %d", len(stats.CloneInfos))
	}

	if len(stats.CloneErrors) != 0 {
		t.Errorf("Expected no clone errors for wiki, got %d", len(stats.CloneErrors))
	}
}

func TestRepositoryProcessor_ProcessRepository_BackupMode(t *testing.T) {
	defer UnsetEnv("GHORG_")()
	os.Setenv("GHORG_BACKUP", "true")

	// Set up temporary directory with existing repo
	dir, err := os.MkdirTemp("", "ghorg_test_backup_mode")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	outputDirAbsolutePath = dir

	// Create existing repo directory
	repoDir := filepath.Join(dir, "test-repo")
	err = os.MkdirAll(repoDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	mockGit := NewExtendedMockGit()
	processor := NewRepositoryProcessor(mockGit)

	repo := scm.Repo{
		Name:        "test-repo",
		URL:         "https://github.com/org/test-repo",
		CloneBranch: "main",
		HostPath:    repoDir,
	}

	repoNameWithCollisions := make(map[string]bool)
	processor.ProcessRepository(&repo, repoNameWithCollisions, false, "test-repo", 0)

	stats := processor.GetStats()
	if stats.UpdateRemoteCount != 1 {
		t.Errorf("Expected update remote count to be 1, got %d", stats.UpdateRemoteCount)
	}
}

func TestRepositoryProcessor_ProcessRepository_NoCleanMode(t *testing.T) {
	defer UnsetEnv("GHORG_")()
	os.Setenv("GHORG_NO_CLEAN", "true")

	// Set up temporary directory with existing repo
	dir, err := os.MkdirTemp("", "ghorg_test_no_clean_mode")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	outputDirAbsolutePath = dir

	// Create existing repo directory
	repoDir := filepath.Join(dir, "test-repo")
	err = os.MkdirAll(repoDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	mockGit := NewExtendedMockGit()
	processor := NewRepositoryProcessor(mockGit)

	repo := scm.Repo{
		Name:        "test-repo",
		URL:         "https://github.com/org/test-repo",
		CloneBranch: "main",
		HostPath:    repoDir,
	}

	repoNameWithCollisions := make(map[string]bool)
	processor.ProcessRepository(&repo, repoNameWithCollisions, false, "test-repo", 0)

	stats := processor.GetStats()
	// In no-clean mode, we still increment pulled count
	if stats.PulledCount != 1 {
		t.Errorf("Expected pulled count to be 1, got %d", stats.PulledCount)
	}
}

func TestRepositoryProcessor_ProcessRepository_NameCollisions(t *testing.T) {
	defer UnsetEnv("GHORG_")()

	dir, err := os.MkdirTemp("", "ghorg_test_name_collisions")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	outputDirAbsolutePath = dir

	mockGit := NewExtendedMockGit()
	processor := NewRepositoryProcessor(mockGit)

	repo := scm.Repo{
		Name:        "test-repo",
		URL:         "https://github.com/org/test-repo",
		CloneBranch: "main",
		Path:        "group/subgroup/test-repo",
	}

	repoNameWithCollisions := map[string]bool{
		"test-repo": true,
	}

	processor.ProcessRepository(&repo, repoNameWithCollisions, true, "test-repo", 1)

	// Check that the repo was processed despite collisions
	stats := processor.GetStats()
	if stats.CloneCount != 1 {
		t.Errorf("Expected clone count to be 1, got %d", stats.CloneCount)
	}

	// The host path should be modified due to collision handling
	expectedPath := filepath.Join(outputDirAbsolutePath, "group_subgroup_test-repo")
	if repo.HostPath != expectedPath {
		t.Errorf("Expected host path to be modified for collisions, got %s", repo.HostPath)
	}
}

func TestRepositoryProcessor_ProcessRepository_CrossPlatformPaths(t *testing.T) {
	defer UnsetEnv("GHORG_")()

	dir, err := os.MkdirTemp("", "ghorg_test_cross_platform")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	outputDirAbsolutePath = dir

	mockGit := NewExtendedMockGit()
	processor := NewRepositoryProcessor(mockGit)

	// Test with forward slashes (Unix-style)
	repoUnix := scm.Repo{
		Name:        "test-repo",
		URL:         "https://github.com/org/test-repo",
		CloneBranch: "main",
		Path:        "group/subgroup/test-repo",
	}

	// Test with backslashes (Windows-style)
	repoWindows := scm.Repo{
		Name:        "test-repo2",
		URL:         "https://github.com/org/test-repo2",
		CloneBranch: "main",
		Path:        "group\\subgroup\\test-repo2",
	}

	repoNameWithCollisions := map[string]bool{
		"test-repo":  true,
		"test-repo2": true,
	}

	// Process Unix-style path
	processor.ProcessRepository(&repoUnix, repoNameWithCollisions, true, "test-repo", 0)
	expectedUnixPath := filepath.Join(outputDirAbsolutePath, "group_subgroup_test-repo")
	if repoUnix.HostPath != expectedUnixPath {
		t.Errorf("Expected Unix-style path to be %s, got %s", expectedUnixPath, repoUnix.HostPath)
	}

	// Process Windows-style path
	processor.ProcessRepository(&repoWindows, repoNameWithCollisions, true, "test-repo2", 1)
	expectedWindowsPath := filepath.Join(outputDirAbsolutePath, "group_subgroup_test-repo2")
	if repoWindows.HostPath != expectedWindowsPath {
		t.Errorf("Expected Windows-style path to be %s, got %s", expectedWindowsPath, repoWindows.HostPath)
	}

	stats := processor.GetStats()
	if stats.CloneCount != 2 {
		t.Errorf("Expected clone count to be 2, got %d", stats.CloneCount)
	}
}

func TestRepositoryProcessor_ProcessRepository_GitLabSnippets(t *testing.T) {
	defer UnsetEnv("GHORG_")()

	dir, err := os.MkdirTemp("", "ghorg_test_gitlab_snippets")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	outputDirAbsolutePath = dir

	mockGit := NewExtendedMockGit()
	processor := NewRepositoryProcessor(mockGit)

	// Test regular snippet
	repo := scm.Repo{
		Name:            "test-repo",
		URL:             "https://gitlab.com/org/test-repo",
		CloneBranch:     "main",
		IsGitLabSnippet: true,
		GitLabSnippetInfo: scm.GitLabSnippet{
			Title:     "My Snippet",
			ID:        "123",
			URLOfRepo: "https://gitlab.com/org/test-repo.git",
		},
	}

	repoNameWithCollisions := make(map[string]bool)
	processor.ProcessRepository(&repo, repoNameWithCollisions, false, "test-repo", 0)

	expectedPath := filepath.Join(outputDirAbsolutePath, "test-repo.snippets", "My Snippet-123")
	if repo.HostPath != expectedPath {
		t.Errorf("Expected host path %s, got %s", expectedPath, repo.HostPath)
	}

	// Test root level snippet
	rootSnippetRepo := scm.Repo{
		Name:                     "root-snippet",
		URL:                      "https://gitlab.com/snippets/456",
		CloneBranch:              "main",
		IsGitLabSnippet:          true,
		IsGitLabRootLevelSnippet: true,
		GitLabSnippetInfo: scm.GitLabSnippet{
			Title: "Root Snippet",
			ID:    "456",
		},
	}

	processor.ProcessRepository(&rootSnippetRepo, repoNameWithCollisions, false, "root-snippet", 0)

	expectedRootPath := filepath.Join(outputDirAbsolutePath, "_ghorg_root_level_snippets", "Root Snippet-456")
	if rootSnippetRepo.HostPath != expectedRootPath {
		t.Errorf("Expected host path %s, got %s", expectedRootPath, rootSnippetRepo.HostPath)
	}

	stats := processor.GetStats()
	if stats.CloneCount != 2 {
		t.Errorf("Expected clone count to be 2, got %d", stats.CloneCount)
	}
}

func TestRepositoryProcessor_GetStats(t *testing.T) {
	mockGit := NewExtendedMockGit()
	processor := NewRepositoryProcessor(mockGit)

	// Add some stats manually
	processor.addError("test error")
	processor.addInfo("test info")

	stats := processor.GetStats()

	if len(stats.CloneErrors) != 1 {
		t.Errorf("Expected 1 clone error, got %d", len(stats.CloneErrors))
	}

	if stats.CloneErrors[0] != "test error" {
		t.Errorf("Expected error message 'test error', got '%s'", stats.CloneErrors[0])
	}

	if len(stats.CloneInfos) != 1 {
		t.Errorf("Expected 1 clone info, got %d", len(stats.CloneInfos))
	}

	if stats.CloneInfos[0] != "test info" {
		t.Errorf("Expected info message 'test info', got '%s'", stats.CloneInfos[0])
	}
}

func TestRepositoryProcessor_ThreadSafety(t *testing.T) {
	defer UnsetEnv("GHORG_")()

	dir, err := os.MkdirTemp("", "ghorg_test_thread_safety")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	outputDirAbsolutePath = dir

	mockGit := NewExtendedMockGit()
	processor := NewRepositoryProcessor(mockGit)

	// Simulate concurrent access
	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			processor.addError("error " + string(rune(index)))
			processor.addInfo("info " + string(rune(index)))
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	stats := processor.GetStats()
	if len(stats.CloneErrors) != numGoroutines {
		t.Errorf("Expected %d clone errors, got %d", numGoroutines, len(stats.CloneErrors))
	}

	if len(stats.CloneInfos) != numGoroutines {
		t.Errorf("Expected %d clone infos, got %d", numGoroutines, len(stats.CloneInfos))
	}
}
