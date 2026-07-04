package configs_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/gabrie30/ghorg/configs"
)

func TestVerifyTokenSet(t *testing.T) {

	t.Run("When cloning github", func(tt *testing.T) {
		_ = os.Setenv("GHORG_SCM_TYPE", "github")
		_ = os.Setenv("GHORG_GITHUB_TOKEN", "")

		err := configs.VerifyTokenSet()
		if err != configs.ErrNoGitHubToken {
			tt.Errorf("Expected ErrNoGitHubTokenError, got: %v", err)
		}

	})

	t.Run("When cloning gitlab", func(tt *testing.T) {
		_ = os.Setenv("GHORG_SCM_TYPE", "gitlab")
		_ = os.Setenv("GHORG_GITLAB_TOKEN", "")

		err := configs.VerifyTokenSet()
		if err != configs.ErrNoGitLabToken {
			tt.Errorf("Expected ErrNoGitLabTokenError, got: %v", err)
		}

	})

	t.Run("When cloning bitbucket with no username", func(tt *testing.T) {
		_ = os.Setenv("GHORG_SCM_TYPE", "bitbucket")
		_ = os.Setenv("GHORG_BITBUCKET_USERNAME", "")
		_ = os.Setenv("GHORG_BITBUCKET_APP_PASSWORD", "12345678912345678901")
		_ = os.Setenv("GHORG_BITBUCKET_API_TOKEN", "")
		_ = os.Setenv("GHORG_BITBUCKET_API_EMAIL", "")
		_ = os.Setenv("GHORG_BITBUCKET_OAUTH_TOKEN", "")
		err := configs.VerifyTokenSet()
		if err != configs.ErrNoBitbucketUsername {
			tt.Errorf("Expected ErrNoBitbucketUsername, got: %v", err)
		}

	})

	t.Run("When cloning bitbucket with username but no app password", func(tt *testing.T) {
		_ = os.Setenv("GHORG_SCM_TYPE", "bitbucket")
		_ = os.Setenv("GHORG_BITBUCKET_USERNAME", "bitbucketuser")
		_ = os.Setenv("GHORG_BITBUCKET_APP_PASSWORD", "")
		_ = os.Setenv("GHORG_BITBUCKET_API_TOKEN", "")
		_ = os.Setenv("GHORG_BITBUCKET_API_EMAIL", "")
		_ = os.Setenv("GHORG_BITBUCKET_OAUTH_TOKEN", "")
		err := configs.VerifyTokenSet()
		if err != configs.ErrNoBitbucketAppPassword {
			tt.Errorf("Expected ErrNoBitbucketAppPassword, got: %v", err)
		}

	})

	t.Run("When cloning bitbucket with API token but no email or username", func(tt *testing.T) {
		_ = os.Setenv("GHORG_SCM_TYPE", "bitbucket")
		_ = os.Setenv("GHORG_BITBUCKET_USERNAME", "")
		_ = os.Setenv("GHORG_BITBUCKET_APP_PASSWORD", "")
		_ = os.Setenv("GHORG_BITBUCKET_API_TOKEN", "test_api_token")
		_ = os.Setenv("GHORG_BITBUCKET_API_EMAIL", "")
		_ = os.Setenv("GHORG_BITBUCKET_OAUTH_TOKEN", "")
		err := configs.VerifyTokenSet()
		if err != configs.ErrNoBitbucketAPIEmail {
			tt.Errorf("Expected ErrNoBitbucketAPIEmail, got: %v", err)
		}

	})

	t.Run("When cloning bitbucket with API token and email", func(tt *testing.T) {
		_ = os.Setenv("GHORG_SCM_TYPE", "bitbucket")
		_ = os.Setenv("GHORG_BITBUCKET_USERNAME", "")
		_ = os.Setenv("GHORG_BITBUCKET_APP_PASSWORD", "")
		_ = os.Setenv("GHORG_BITBUCKET_API_TOKEN", "test_api_token")
		_ = os.Setenv("GHORG_BITBUCKET_API_EMAIL", "test@example.com")
		_ = os.Setenv("GHORG_BITBUCKET_OAUTH_TOKEN", "")
		err := configs.VerifyTokenSet()
		if err != nil {
			tt.Errorf("Expected no error, got: %v", err)
		}

	})

	t.Run("When cloning bitbucket with API token and username fallback", func(tt *testing.T) {
		_ = os.Setenv("GHORG_SCM_TYPE", "bitbucket")
		_ = os.Setenv("GHORG_BITBUCKET_USERNAME", "test@example.com")
		_ = os.Setenv("GHORG_BITBUCKET_APP_PASSWORD", "")
		_ = os.Setenv("GHORG_BITBUCKET_API_TOKEN", "test_api_token")
		_ = os.Setenv("GHORG_BITBUCKET_API_EMAIL", "")
		_ = os.Setenv("GHORG_BITBUCKET_OAUTH_TOKEN", "")
		err := configs.VerifyTokenSet()
		if err != nil {
			tt.Errorf("Expected no error, got: %v", err)
		}

	})

	t.Run("When cloning bitbucket with OAuth token", func(tt *testing.T) {
		_ = os.Setenv("GHORG_SCM_TYPE", "bitbucket")
		_ = os.Setenv("GHORG_BITBUCKET_USERNAME", "")
		_ = os.Setenv("GHORG_BITBUCKET_APP_PASSWORD", "")
		_ = os.Setenv("GHORG_BITBUCKET_API_TOKEN", "")
		_ = os.Setenv("GHORG_BITBUCKET_API_EMAIL", "")
		_ = os.Setenv("GHORG_BITBUCKET_OAUTH_TOKEN", "oauth_token")
		err := configs.VerifyTokenSet()
		if err != nil {
			tt.Errorf("Expected no error, got: %v", err)
		}

	})
}

func TestVerifyConfigsSetCorrectly(t *testing.T) {

	t.Run("When unsupported scm", func(tt *testing.T) {
		_ = os.Setenv("GHORG_CLONE_TYPE", "org")
		_ = os.Setenv("GHORG_CLONE_PROTOCOL", "ssh")

		_ = os.Setenv("GHORG_SCM_TYPE", "githubz")

		err := configs.VerifyConfigsSetCorrectly()
		if err != configs.ErrIncorrectScmType {
			tt.Errorf("Expected ErrIncorrectScmType, got: %v", err)
		}

	})

	t.Run("When unsupported clone type", func(tt *testing.T) {
		_ = os.Setenv("GHORG_SCM_TYPE", "github")
		_ = os.Setenv("GHORG_CLONE_PROTOCOL", "ssh")

		_ = os.Setenv("GHORG_CLONE_TYPE", "bot")

		err := configs.VerifyConfigsSetCorrectly()
		if err != configs.ErrIncorrectCloneType {
			tt.Errorf("Expected ErrIncorrectCloneType, got: %v", err)
		}

	})

	t.Run("When unsupported protocol", func(tt *testing.T) {
		_ = os.Setenv("GHORG_SCM_TYPE", "github")
		_ = os.Setenv("GHORG_CLONE_TYPE", "org")

		_ = os.Setenv("GHORG_CLONE_PROTOCOL", "ftp")

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
