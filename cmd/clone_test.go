package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/gabrie30/ghorg/scm"
)

func TestShouldLowerRegularString(t *testing.T) {

	upperName := "RepoName"
	parseParentFolder([]string{upperName})

	if parentFolder != "reponame" {
		t.Errorf("Wrong folder name, expected: %s, got: %s", upperName, parentFolder)
	}
}

func TestShouldNotChangeLowerCasedRegularString(t *testing.T) {

	lowerName := "repo_name"
	parseParentFolder([]string{lowerName})

	if parentFolder != "repo_name" {
		t.Errorf("Wrong folder name, expected: %s, got: %s", lowerName, parentFolder)
	}
}

func TestReplaceDashWithUnderscore(t *testing.T) {

	want := "repo-name"
	lowerName := "repo-name"
	parseParentFolder([]string{lowerName})

	if parentFolder != want {
		t.Errorf("Wrong folder name, expected: %s, got: %s", want, parentFolder)
	}
}

func TestShouldNotChangeNonLettersString(t *testing.T) {

	numberName := "1234567_8"
	parseParentFolder([]string{numberName})

	if parentFolder != "1234567_8" {
		t.Errorf("Wrong folder name, expected: %s, got: %s", numberName, parentFolder)
	}
}

type MockGitClient struct{}

func NewMockGit() MockGitClient {
	return MockGitClient{}
}

func (g MockGitClient) Clone(repo scm.Repo) error {
	_, err := ioutil.TempDir(os.Getenv("GHORG_ABSOLUTE_PATH_TO_CLONE_TO"), "ghorg_test_repo")
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (g MockGitClient) SetOrigin(repo scm.Repo) error {
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
	dir, err := ioutil.TempDir(".", "ghorg_tests")
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
	got, _ := ioutil.ReadDir(dir)
	expected := len(testRepos)
	if len(got) != expected {
		t.Errorf("Wrong number of repos in clone, expected: %v, got: %v", expected, got)
	}
}
