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
	shouldFailClone           bool
	shouldFailCheckout        bool
	shouldFailSetOrigin       bool
	shouldReturnEmptyRepo     bool
	shouldReturnDirtyStatus   bool
	shouldReturnUnpushedCommits bool
	preCommitCount            int
	postCommitCount           int
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

func (g *ExtendedMockGitClient) ShortStatus(repo scm.Repo) (string, error) {
	if g.shouldReturnDirtyStatus {
		return " M modified_file.go", nil
	}
	return "", nil
}

func (g *ExtendedMockGitClient) RevListCompare(repo scm.Repo, ref1 string, ref2 string) (string, error) {
	if g.shouldReturnUnpushedCommits {
		return "abc123\ndef456", nil
	}
	return "", nil
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

func TestRepositoryProcessor_ProcessRepository_NoCleanModeWithFetchAllDisabled(t *testing.T) {
	defer UnsetEnv("GHORG_")()
	os.Setenv("GHORG_NO_CLEAN", "true")
	os.Setenv("GHORG_FETCH_ALL", "false")

	// Set up temporary directory with existing repo
	dir, err := os.MkdirTemp("", "ghorg_test_no_clean_fetch_all_disabled")
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
	// In no-clean mode with fetch-all disabled, we should still process successfully
	if stats.PulledCount != 1 {
		t.Errorf("Expected pulled count to be 1, got %d", stats.PulledCount)
	}
	// Should not have any errors since fetch-all is skipped when disabled
	if len(stats.CloneErrors) != 0 {
		t.Errorf("Expected no errors when FETCH_ALL is disabled, got %d errors: %v", len(stats.CloneErrors), stats.CloneErrors)
	}
}

func TestRepositoryProcessor_ProcessRepository_NoCleanModeWithFetchAllEnabled(t *testing.T) {
	defer UnsetEnv("GHORG_")()
	os.Setenv("GHORG_NO_CLEAN", "true")
	os.Setenv("GHORG_FETCH_ALL", "true")

	// Set up temporary directory with existing repo
	dir, err := os.MkdirTemp("", "ghorg_test_no_clean_fetch_all_enabled")
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
	// In no-clean mode with fetch-all enabled, we should still process successfully
	if stats.PulledCount != 1 {
		t.Errorf("Expected pulled count to be 1, got %d", stats.PulledCount)
	}
	// Should not have any errors since fetch-all is enabled and mocked
	if len(stats.CloneErrors) != 0 {
		t.Errorf("Expected no errors when FETCH_ALL is enabled, got %d errors: %v", len(stats.CloneErrors), stats.CloneErrors)
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

func TestRepositoryProcessor_SetTotalDuration(t *testing.T) {
	mockGit := NewExtendedMockGit()
	processor := NewRepositoryProcessor(mockGit)

	// Test setting timing
	processor.SetTotalDuration(42)

	stats := processor.GetStats()
	if stats.TotalDurationSeconds != 42 {
		t.Errorf("Expected total duration to be 42, got %d", stats.TotalDurationSeconds)
	}
}

func TestRepositoryProcessor_SetTotalDuration_ThreadSafety(t *testing.T) {
	mockGit := NewExtendedMockGit()
	processor := NewRepositoryProcessor(mockGit)

	// Test concurrent timing updates
	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			processor.SetTotalDuration(index * 10)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// The final value should be one of the set values (race condition, but still valid)
	stats := processor.GetStats()
	validValues := make(map[int]bool)
	for i := 0; i < numGoroutines; i++ {
		validValues[i*10] = true
	}

	if !validValues[stats.TotalDurationSeconds] {
		t.Errorf("Expected total duration to be one of the set values, got %d", stats.TotalDurationSeconds)
	}
}

func TestCloneStats_NewStruct(t *testing.T) {
	stats := CloneStats{
		CloneCount:           5,
		PulledCount:          3,
		UpdateRemoteCount:    2,
		NewCommits:           10,
		UntouchedPrunes:      1,
		TotalDurationSeconds: 120,
		CloneInfos:           []string{"info1", "info2"},
		CloneErrors:          []string{"error1"},
	}

	// Verify all fields are properly set
	if stats.CloneCount != 5 {
		t.Errorf("Expected CloneCount to be 5, got %d", stats.CloneCount)
	}
	if stats.PulledCount != 3 {
		t.Errorf("Expected PulledCount to be 3, got %d", stats.PulledCount)
	}
	if stats.UpdateRemoteCount != 2 {
		t.Errorf("Expected UpdateRemoteCount to be 2, got %d", stats.UpdateRemoteCount)
	}
	if stats.NewCommits != 10 {
		t.Errorf("Expected NewCommits to be 10, got %d", stats.NewCommits)
	}
	if stats.UntouchedPrunes != 1 {
		t.Errorf("Expected UntouchedPrunes to be 1, got %d", stats.UntouchedPrunes)
	}
	if stats.TotalDurationSeconds != 120 {
		t.Errorf("Expected TotalDurationSeconds to be 120, got %d", stats.TotalDurationSeconds)
	}
	if len(stats.CloneInfos) != 2 {
		t.Errorf("Expected 2 CloneInfos, got %d", len(stats.CloneInfos))
	}
	if len(stats.CloneErrors) != 1 {
		t.Errorf("Expected 1 CloneError, got %d", len(stats.CloneErrors))
	}
}

func TestRepositoryProcessor_ProtectLocal_SkipsDirtyRepo(t *testing.T) {
	defer UnsetEnv("GHORG_")()
	os.Setenv("GHORG_PROTECT_LOCAL", "true")

	// Set up temporary directory with existing repo
	dir, err := os.MkdirTemp("", "ghorg_test_protect_local_dirty")
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
	mockGit.shouldReturnDirtyStatus = true
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
	// Should not be pulled because it has uncommitted changes
	if stats.PulledCount != 0 {
		t.Errorf("Expected pulled count to be 0, got %d", stats.PulledCount)
	}
	// Should be counted as protected
	if stats.ProtectedCount != 1 {
		t.Errorf("Expected protected count to be 1, got %d", stats.ProtectedCount)
	}
	// Verify GetProtectedRepos returns the repo
	protectedRepos := processor.GetProtectedRepos()
	if len(protectedRepos) != 1 {
		t.Errorf("Expected 1 protected repo, got %d", len(protectedRepos))
	}
}

func TestRepositoryProcessor_ProtectLocal_SkipsUnpushedCommits(t *testing.T) {
	defer UnsetEnv("GHORG_")()
	os.Setenv("GHORG_PROTECT_LOCAL", "true")

	// Set up temporary directory with existing repo
	dir, err := os.MkdirTemp("", "ghorg_test_protect_local_unpushed")
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
	mockGit.shouldReturnUnpushedCommits = true
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
	// Should not be pulled because it has unpushed commits
	if stats.PulledCount != 0 {
		t.Errorf("Expected pulled count to be 0, got %d", stats.PulledCount)
	}
	// Should be counted as protected
	if stats.ProtectedCount != 1 {
		t.Errorf("Expected protected count to be 1, got %d", stats.ProtectedCount)
	}
}

func TestRepositoryProcessor_ProtectLocal_AllowsCleanRepo(t *testing.T) {
	defer UnsetEnv("GHORG_")()
	os.Setenv("GHORG_PROTECT_LOCAL", "true")

	// Set up temporary directory with existing repo
	dir, err := os.MkdirTemp("", "ghorg_test_protect_local_clean")
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
	// Neither dirty status nor unpushed commits - repo is clean
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
	// Should be pulled because repo is clean
	if stats.PulledCount != 1 {
		t.Errorf("Expected pulled count to be 1, got %d", stats.PulledCount)
	}
	// Should not be counted as protected
	if stats.ProtectedCount != 0 {
		t.Errorf("Expected protected count to be 0, got %d", stats.ProtectedCount)
	}
}

func TestRepositoryProcessor_ProtectLocal_BypassedInBackupMode(t *testing.T) {
	defer UnsetEnv("GHORG_")()
	os.Setenv("GHORG_PROTECT_LOCAL", "true")
	os.Setenv("GHORG_BACKUP", "true")

	// Set up temporary directory with existing repo
	dir, err := os.MkdirTemp("", "ghorg_test_protect_local_backup")
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
	mockGit.shouldReturnDirtyStatus = true // Would normally be skipped
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
	// Backup mode is safe, so protect-local check is bypassed
	if stats.UpdateRemoteCount != 1 {
		t.Errorf("Expected update remote count to be 1, got %d", stats.UpdateRemoteCount)
	}
	// Should not be counted as protected since backup mode is already safe
	if stats.ProtectedCount != 0 {
		t.Errorf("Expected protected count to be 0, got %d", stats.ProtectedCount)
	}
}

func TestRepositoryProcessor_ProtectLocal_BypassedInNoCleanMode(t *testing.T) {
	defer UnsetEnv("GHORG_")()
	os.Setenv("GHORG_PROTECT_LOCAL", "true")
	os.Setenv("GHORG_NO_CLEAN", "true")

	// Set up temporary directory with existing repo
	dir, err := os.MkdirTemp("", "ghorg_test_protect_local_no_clean")
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
	mockGit.shouldReturnDirtyStatus = true // Would normally be skipped
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
	// No-clean mode is safe, so protect-local check is bypassed
	if stats.PulledCount != 1 {
		t.Errorf("Expected pulled count to be 1, got %d", stats.PulledCount)
	}
	// Should not be counted as protected since no-clean mode is already safe
	if stats.ProtectedCount != 0 {
		t.Errorf("Expected protected count to be 0, got %d", stats.ProtectedCount)
	}
}

func TestRepositoryProcessor_ProtectLocal_DisabledByDefault(t *testing.T) {
	defer UnsetEnv("GHORG_")()
	// GHORG_PROTECT_LOCAL is NOT set

	// Set up temporary directory with existing repo
	dir, err := os.MkdirTemp("", "ghorg_test_protect_local_disabled")
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
	mockGit.shouldReturnDirtyStatus = true // Would be skipped if protect-local was enabled
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
	// Should be pulled because protect-local is not enabled
	if stats.PulledCount != 1 {
		t.Errorf("Expected pulled count to be 1, got %d", stats.PulledCount)
	}
	// Should not be counted as protected
	if stats.ProtectedCount != 0 {
		t.Errorf("Expected protected count to be 0, got %d", stats.ProtectedCount)
	}
}

func TestRepositoryProcessor_ProtectLocal_NewRepoNotAffected(t *testing.T) {
	defer UnsetEnv("GHORG_")()
	os.Setenv("GHORG_PROTECT_LOCAL", "true")

	// Set up temporary directory (no existing repo)
	dir, err := os.MkdirTemp("", "ghorg_test_protect_local_new_repo")
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
	// New repos should be cloned regardless of protect-local
	if stats.CloneCount != 1 {
		t.Errorf("Expected clone count to be 1, got %d", stats.CloneCount)
	}
	// Should not be counted as protected
	if stats.ProtectedCount != 0 {
		t.Errorf("Expected protected count to be 0, got %d", stats.ProtectedCount)
	}
}

func TestRepositoryProcessor_HasLocalChanges(t *testing.T) {
	mockGit := NewExtendedMockGit()
	processor := NewRepositoryProcessor(mockGit)

	repo := scm.Repo{
		Name:        "test-repo",
		URL:         "https://github.com/org/test-repo",
		CloneBranch: "main",
		HostPath:    "/tmp/test-repo",
	}

	// Test clean repo
	hasChanges, reason, err := processor.hasLocalChanges(&repo)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if hasChanges {
		t.Error("Expected hasChanges to be false for clean repo")
	}
	if reason != "" {
		t.Errorf("Expected empty reason, got %s", reason)
	}

	// Test dirty status
	mockGit.shouldReturnDirtyStatus = true
	hasChanges, reason, err = processor.hasLocalChanges(&repo)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !hasChanges {
		t.Error("Expected hasChanges to be true for dirty repo")
	}
	if reason != "uncommitted changes" {
		t.Errorf("Expected reason to be 'uncommitted changes', got %s", reason)
	}

	// Test unpushed commits (reset dirty status first)
	mockGit.shouldReturnDirtyStatus = false
	mockGit.shouldReturnUnpushedCommits = true
	hasChanges, reason, err = processor.hasLocalChanges(&repo)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !hasChanges {
		t.Error("Expected hasChanges to be true for repo with unpushed commits")
	}
	if reason != "unpushed commits" {
		t.Errorf("Expected reason to be 'unpushed commits', got %s", reason)
	}
}

func TestRepositoryProcessor_AddProtected(t *testing.T) {
	mockGit := NewExtendedMockGit()
	processor := NewRepositoryProcessor(mockGit)

	// Add protected repos
	processor.addProtected("/tmp/repo1")
	processor.addProtected("/tmp/repo2")

	stats := processor.GetStats()
	if stats.ProtectedCount != 2 {
		t.Errorf("Expected protected count to be 2, got %d", stats.ProtectedCount)
	}

	protectedRepos := processor.GetProtectedRepos()
	if len(protectedRepos) != 2 {
		t.Errorf("Expected 2 protected repos, got %d", len(protectedRepos))
	}
	if protectedRepos[0] != "/tmp/repo1" {
		t.Errorf("Expected first protected repo to be '/tmp/repo1', got %s", protectedRepos[0])
	}
	if protectedRepos[1] != "/tmp/repo2" {
		t.Errorf("Expected second protected repo to be '/tmp/repo2', got %s", protectedRepos[1])
	}
}
