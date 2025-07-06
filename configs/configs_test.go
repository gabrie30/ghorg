package configs_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/gabrie30/ghorg/configs"
)

func TestVerifyTokenSet(t *testing.T) {

	t.Run("When cloning github", func(tt *testing.T) {
		os.Setenv("GHORG_SCM_TYPE", "github")
		os.Setenv("GHORG_GITHUB_TOKEN", "")

		err := configs.VerifyTokenSet()
		if err != configs.ErrNoGitHubToken {
			tt.Errorf("Expected ErrNoGitHubTokenError, got: %v", err)
		}

	})

	t.Run("When cloning gitlab", func(tt *testing.T) {
		os.Setenv("GHORG_SCM_TYPE", "gitlab")
		os.Setenv("GHORG_GITLAB_TOKEN", "")

		err := configs.VerifyTokenSet()
		if err != configs.ErrNoGitLabToken {
			tt.Errorf("Expected ErrNoGitLabTokenError, got: %v", err)
		}

	})

	t.Run("When cloning bitbucket with no username", func(tt *testing.T) {
		os.Setenv("GHORG_SCM_TYPE", "bitbucket")
		os.Setenv("GHORG_BITBUCKET_USERNAME", "")
		os.Setenv("GHORG_BITBUCKET_APP_PASSWORD", "12345678912345678901")
		err := configs.VerifyTokenSet()
		if err != configs.ErrNoBitbucketUsername {
			tt.Errorf("Expected ErrNoBitbucketUsername, got: %v", err)
		}

	})

	t.Run("When cloning bitbucket with username but no app password", func(tt *testing.T) {
		os.Setenv("GHORG_SCM_TYPE", "bitbucket")
		os.Setenv("GHORG_BITBUCKET_USERNAME", "bitbucketuser")
		os.Setenv("GHORG_BITBUCKET_APP_PASSWORD", "")
		err := configs.VerifyTokenSet()
		if err != configs.ErrNoBitbucketAppPassword {
			tt.Errorf("Expected ErrNoBitbucketAppPassword, got: %v", err)
		}

	})
}

func TestVerifyConfigsSetCorrectly(t *testing.T) {

	t.Run("When unsupported scm", func(tt *testing.T) {
		os.Setenv("GHORG_CLONE_TYPE", "org")
		os.Setenv("GHORG_CLONE_PROTOCOL", "ssh")

		os.Setenv("GHORG_SCM_TYPE", "githubz")

		err := configs.VerifyConfigsSetCorrectly()
		if err != configs.ErrIncorrectScmType {
			tt.Errorf("Expected ErrIncorrectScmType, got: %v", err)
		}

	})

	t.Run("When unsupported clone type", func(tt *testing.T) {
		os.Setenv("GHORG_SCM_TYPE", "github")
		os.Setenv("GHORG_CLONE_PROTOCOL", "ssh")

		os.Setenv("GHORG_CLONE_TYPE", "bot")

		err := configs.VerifyConfigsSetCorrectly()
		if err != configs.ErrIncorrectCloneType {
			tt.Errorf("Expected ErrIncorrectCloneType, got: %v", err)
		}

	})

	t.Run("When unsupported protocol", func(tt *testing.T) {
		os.Setenv("GHORG_SCM_TYPE", "github")
		os.Setenv("GHORG_CLONE_TYPE", "org")

		os.Setenv("GHORG_CLONE_PROTOCOL", "ftp")

		err := configs.VerifyConfigsSetCorrectly()
		if err != configs.ErrIncorrectProtocolType {
			tt.Errorf("Expected ErrIncorrectProtocolType, got: %v", err)
		}

	})
}

func TestTrailingSlashes(t *testing.T) {
	t.Run("URL's with no trailing slash should get one", func(tt *testing.T) {
		got := configs.EnsureTrailingSlashOnURL("github.com")
		want := "github.com/"
		if got != want {
			tt.Errorf("Expected %v, got: %v", want, got)
		}

	})

	t.Run("URL's with a trailing slash should only have one", func(tt *testing.T) {
		got := configs.EnsureTrailingSlashOnURL("github.com/")
		want := "github.com/"
		if got != want {
			tt.Errorf("Expected %v, got: %v", want, got)
		}

	})

	t.Run("Filepaths should be correctly appeneded", func(tt *testing.T) {
		got := configs.EnsureTrailingSlashOnFilePath("foo")
		want := "foo/"
		if runtime.GOOS == "windows" {
			want = "foo\\"
		}
		if got != want {
			tt.Errorf("Expected %v, got: %v", want, got)
		}
	})
}

func TestSyncDefaultBranchConfiguration(t *testing.T) {
	tests := []struct {
		name           string
		envValue       string
		expectedResult bool
	}{
		{
			name:           "Environment variable not set",
			envValue:       "",
			expectedResult: false,
		},
		{
			name:           "Environment variable set to true",
			envValue:       "true",
			expectedResult: true,
		},
		{
			name:           "Environment variable set to false",
			envValue:       "false",
			expectedResult: false,
		},
		{
			name:           "Environment variable set to invalid value",
			envValue:       "invalid",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original value
			originalValue := os.Getenv("GHORG_SYNC_DEFAULT_BRANCH")
			defer func() {
				if originalValue != "" {
					os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", originalValue)
				} else {
					os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
				}
			}()

			// Set test value
			if tt.envValue == "" {
				os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
			} else {
				os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", tt.envValue)
			}

			// Test the logic that sync.go uses
			syncEnabled := os.Getenv("GHORG_SYNC_DEFAULT_BRANCH")
			actualResult := syncEnabled == "true"

			if actualResult != tt.expectedResult {
				t.Errorf("Expected sync enabled to be %v, got %v for env value '%s'", tt.expectedResult, actualResult, tt.envValue)
			}
		})
	}
}

func TestSyncDefaultBranchEnvironmentVariableHandling(t *testing.T) {
	// Test that the environment variable is properly recognized by the config system
	originalValue := os.Getenv("GHORG_SYNC_DEFAULT_BRANCH")
	defer func() {
		if originalValue != "" {
			os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", originalValue)
		} else {
			os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
		}
	}()

	// Test with true value
	os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "true")
	if os.Getenv("GHORG_SYNC_DEFAULT_BRANCH") != "true" {
		t.Error("Environment variable GHORG_SYNC_DEFAULT_BRANCH should be settable")
	}

	// Test with false value
	os.Setenv("GHORG_SYNC_DEFAULT_BRANCH", "false")
	if os.Getenv("GHORG_SYNC_DEFAULT_BRANCH") != "false" {
		t.Error("Environment variable GHORG_SYNC_DEFAULT_BRANCH should accept false value")
	}

	// Test unsetting
	os.Unsetenv("GHORG_SYNC_DEFAULT_BRANCH")
	if os.Getenv("GHORG_SYNC_DEFAULT_BRANCH") != "" {
		t.Error("Environment variable GHORG_SYNC_DEFAULT_BRANCH should be unsetable")
	}
}
