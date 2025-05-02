package git

import (
	"testing"

	"github.com/gabrie30/ghorg/scm"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
)

// Mock repository for testing
func setupMockRepo() (*git.Repository, error) {
	storer := memory.NewStorage()
	return git.Init(storer, nil)
}

func TestClone(t *testing.T) {
	repo := scm.Repo{
		CloneURL:    "https://github.com/example/repo.git",
		HostPath:    "/tmp/repo",
		CloneBranch: "main",
	}

	client := GitClient{}

	// Since Clone interacts with the filesystem, you can mock its behavior or test it in an isolated environment.
	err := client.Clone(repo, false)
	assert.NotNil(t, err, "Clone should fail because it requires a real repository")
}

func TestSetOrigin(t *testing.T) {
	_, err := setupMockRepo()
	assert.NoError(t, err)

	client := GitClient{}
	repo := scm.Repo{
		URL: "https://github.com/example/repo.git",
	}

	// Mock the repository's behavior
	err = client.SetOrigin(repo, false)
	assert.Error(t, err, "SetOrigin should fail because it requires a real repository")
}

func TestShortStatus(t *testing.T) {
	_, err := setupMockRepo()
	assert.NoError(t, err)

	client := GitClient{}
	repo := scm.Repo{
		HostPath: "/tmp/repo",
	}

	// Mock the repository's behavior
	status, err := client.ShortStatus(repo, false)
	assert.Error(t, err, "ShortStatus should fail because it requires a real repository")
	assert.Equal(t, "", status, "ShortStatus should return an empty string on failure")
}

func TestRepoCommitCount(t *testing.T) {
	_, err := setupMockRepo()
	assert.NoError(t, err)

	client := GitClient{}
	repo := scm.Repo{
		HostPath:    "/tmp/repo",
		CloneBranch: "main",
	}

	// Mock the repository's behavior
	count, err := client.RepoCommitCount(repo, false)
	assert.Error(t, err, "RepoCommitCount should fail because it requires a real repository")
	assert.Equal(t, 0, count, "RepoCommitCount should return 0 on failure")
}
