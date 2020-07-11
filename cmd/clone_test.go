package cmd

import "testing"

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

	lowerName := "repo-name"
	parseParentFolder([]string{lowerName})

	if parentFolder != "repo_name" {
		t.Errorf("Wrong folder name, expected: %s, got: %s", lowerName, parentFolder)
	}
}

func TestShouldNotChangeNonLettersString(t *testing.T) {

	numberName := "1234567_8"
	parseParentFolder([]string{numberName})

	if parentFolder != "1234567_8" {
		t.Errorf("Wrong folder name, expected: %s, got: %s", numberName, parentFolder)
	}
}
