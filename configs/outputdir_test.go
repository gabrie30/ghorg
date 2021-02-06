package configs

import "testing"

func TestShouldLowerRegularString(t *testing.T) {
	config := &Config{OutputDir: "RepoName"}
	config.parseOutputDir(nil)

	if config.OutputDir != "reponame" {
		t.Errorf("Wrong folder name, expected: %s, got: %s", "reponame", config.OutputDir)
	}
}

func TestShouldNotChangeLowerCasedRegularString(t *testing.T) {
	config := &Config{OutputDir: "repo_name"}
	config.parseOutputDir(nil)

	if config.OutputDir != "repo_name" {
		t.Errorf("Wrong folder name, expected: %s, got: %s", "repo_name", config.OutputDir)
	}
}

func TestReplaceDashWithUnderscore(t *testing.T) {
	config := &Config{OutputDir: "repo-name"}
	config.parseOutputDir(nil)

	if config.OutputDir != "repo_name" {
		t.Errorf("Wrong folder name, expected: %s, got: %s", "repo_name", config.OutputDir)
	}
}

func TestShouldNotChangeNonLettersString(t *testing.T) {
	config := &Config{OutputDir: "1234567_8"}
	config.parseOutputDir(nil)

	if config.OutputDir != "1234567_8" {
		t.Errorf("Wrong folder name, expected: %s, got: %s", "1234567_8", config.OutputDir)
	}
}
