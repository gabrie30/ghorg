package scm

import (
	"os"
	"testing"
)

func TestFilterGitlabGroupByMatchRegex(t *testing.T) {
	testCases := []struct {
		name           string
		regex          string
		groups         []string
		expectedGroups []string
	}{
		{
			name:           "matches specific group names",
			regex:          "^(subgroup-a|subgroup-b)$",
			groups:         []string{"subgroup-a", "subgroup-b", "subgroup-c", "subgroup-d"},
			expectedGroups: []string{"subgroup-a", "subgroup-b"},
		},
		{
			name:           "matches groups with partial regex",
			regex:          "subgroup-a",
			groups:         []string{"subgroup-a", "subgroup-ab", "subgroup-b"},
			expectedGroups: []string{"subgroup-a", "subgroup-ab"},
		},
		{
			name:           "no matches returns empty",
			regex:          "^nonexistent$",
			groups:         []string{"subgroup-a", "subgroup-b"},
			expectedGroups: []string{},
		},
		{
			name:           "matches all groups",
			regex:          ".*",
			groups:         []string{"subgroup-a", "subgroup-b", "subgroup-c"},
			expectedGroups: []string{"subgroup-a", "subgroup-b", "subgroup-c"},
		},
		{
			name:           "case insensitive matching",
			regex:          "(?i:^SUBGROUP-A$)",
			groups:         []string{"subgroup-a", "subgroup-b", "Subgroup-A"},
			expectedGroups: []string{"subgroup-a", "Subgroup-A"},
		},
		{
			name:           "matches numeric group IDs",
			regex:          "^(123|456)$",
			groups:         []string{"123", "456", "789", "101"},
			expectedGroups: []string{"123", "456"},
		},
		{
			name:           "empty groups list",
			regex:          "anything",
			groups:         []string{},
			expectedGroups: []string{},
		},
		{
			name:           "matches groups with path-like names",
			regex:          "my-org/subgroup",
			groups:         []string{"my-org/subgroup-a", "my-org/subgroup-b", "other-org/group-c"},
			expectedGroups: []string{"my-org/subgroup-a", "my-org/subgroup-b"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("GHORG_GITLAB_GROUP_MATCH_REGEX", tc.regex)
			defer os.Unsetenv("GHORG_GITLAB_GROUP_MATCH_REGEX")

			result := filterGitlabGroupByMatchRegex(tc.groups)
			if len(result) == 0 && len(tc.expectedGroups) == 0 {
				return // Both empty, test passes
			}
			if len(result) != len(tc.expectedGroups) {
				t.Errorf("Expected %d groups, got %d. Expected: %v, Got: %v",
					len(tc.expectedGroups), len(result), tc.expectedGroups, result)
				return
			}
			for i, group := range result {
				if group != tc.expectedGroups[i] {
					t.Errorf("Expected group at index %d to be %s, got %s", i, tc.expectedGroups[i], group)
				}
			}
		})
	}
}

func TestFilterGitlabGroupByExcludeMatchRegex(t *testing.T) {
	testCases := []struct {
		name           string
		regex          string
		groups         []string
		expectedGroups []string
	}{
		{
			name:           "excludes specific group names",
			regex:          "^(subgroup-a|subgroup-b)$",
			groups:         []string{"subgroup-a", "subgroup-b", "subgroup-c", "subgroup-d"},
			expectedGroups: []string{"subgroup-c", "subgroup-d"},
		},
		{
			name:           "no exclusions when nothing matches",
			regex:          "^nonexistent$",
			groups:         []string{"subgroup-a", "subgroup-b"},
			expectedGroups: []string{"subgroup-a", "subgroup-b"},
		},
		{
			name:           "excludes all groups",
			regex:          "subgroup",
			groups:         []string{"subgroup-a", "subgroup-b"},
			expectedGroups: []string{},
		},
		{
			name:           "empty groups list",
			regex:          "anything",
			groups:         []string{},
			expectedGroups: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("GHORG_GITLAB_GROUP_EXCLUDE_MATCH_REGEX", tc.regex)
			defer os.Unsetenv("GHORG_GITLAB_GROUP_EXCLUDE_MATCH_REGEX")

			result := filterGitlabGroupByExcludeMatchRegex(tc.groups)
			if len(result) == 0 && len(tc.expectedGroups) == 0 {
				return // Both empty, test passes
			}
			if len(result) != len(tc.expectedGroups) {
				t.Errorf("Expected %d groups, got %d. Expected: %v, Got: %v",
					len(tc.expectedGroups), len(result), tc.expectedGroups, result)
				return
			}
			for i, group := range result {
				if group != tc.expectedGroups[i] {
					t.Errorf("Expected group at index %d to be %s, got %s", i, tc.expectedGroups[i], group)
				}
			}
		})
	}
}

func TestFilterGitlabGroupMatchAndExcludeCombined(t *testing.T) {
	// Test that match and exclude work together:
	// First include only matching, then exclude from those
	t.Run("match then exclude narrows results", func(t *testing.T) {
		groups := []string{"subgroup-a", "subgroup-b", "subgroup-c", "other-group"}

		// First apply match: include only subgroup-* groups
		os.Setenv("GHORG_GITLAB_GROUP_MATCH_REGEX", "^subgroup-")
		defer os.Unsetenv("GHORG_GITLAB_GROUP_MATCH_REGEX")

		matched := filterGitlabGroupByMatchRegex(groups)

		// Then apply exclude: remove subgroup-c from the result
		os.Setenv("GHORG_GITLAB_GROUP_EXCLUDE_MATCH_REGEX", "^subgroup-c$")
		defer os.Unsetenv("GHORG_GITLAB_GROUP_EXCLUDE_MATCH_REGEX")

		result := filterGitlabGroupByExcludeMatchRegex(matched)

		expected := []string{"subgroup-a", "subgroup-b"}
		if len(result) != len(expected) {
			t.Errorf("Expected %d groups, got %d. Expected: %v, Got: %v",
				len(expected), len(result), expected, result)
			return
		}
		for i, group := range result {
			if group != expected[i] {
				t.Errorf("Expected group at index %d to be %s, got %s", i, expected[i], group)
			}
		}
	})
}

func TestFilterGitlabGroupMatchRegexWithAlternation(t *testing.T) {
	// This is the primary use case from the feature request:
	// User wants to include only 3 subgroups out of 20
	t.Run("include only 3 out of many subgroups using alternation", func(t *testing.T) {
		// Simulate 10 subgroups
		groups := []string{
			"subgroup-1", "subgroup-2", "subgroup-3", "subgroup-4", "subgroup-5",
			"subgroup-6", "subgroup-7", "subgroup-8", "subgroup-9", "subgroup-10",
		}

		// Include only 3 specific subgroups
		os.Setenv("GHORG_GITLAB_GROUP_MATCH_REGEX", "^(subgroup-2|subgroup-5|subgroup-8)$")
		defer os.Unsetenv("GHORG_GITLAB_GROUP_MATCH_REGEX")

		result := filterGitlabGroupByMatchRegex(groups)

		expected := []string{"subgroup-2", "subgroup-5", "subgroup-8"}
		if len(result) != len(expected) {
			t.Errorf("Expected %d groups, got %d. Expected: %v, Got: %v",
				len(expected), len(result), expected, result)
			return
		}
		for i, group := range result {
			if group != expected[i] {
				t.Errorf("Expected group at index %d to be %s, got %s", i, expected[i], group)
			}
		}
	})
}
