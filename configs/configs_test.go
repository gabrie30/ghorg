package configs_test

import (
	"os"
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
			tt.Errorf("Expected ErrNoGitHubTokenError, got: %v", err)
		}

	})

	t.Run("When cloning bitbucket with no username", func(tt *testing.T) {
		os.Setenv("GHORG_SCM_TYPE", "bitbucket")
		os.Setenv("GHORG_BITBUCKET_USERNAME", "")
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
