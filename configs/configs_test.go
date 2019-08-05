package configs_test

import (
	"os"
	"testing"

	"github.com/gabrie30/ghorg/configs"
)

func TestDefaultSettings(t *testing.T) {

	branch := os.Getenv("GHORG_BRANCH")
	protocol := os.Getenv("GHORG_CLONE_PROTOCOL")
	scm := os.Getenv("GHORG_SCM_TYPE")
	cloneType := os.Getenv("GHORG_CLONE_TYPE")
	namespace := os.Getenv("GHORG_GITLAB_DEFAULT_NAMESPACE")

	if branch != "master" {
		t.Errorf("Default branch should be master, got: %v", branch)
	}

	if protocol != "https" {
		t.Errorf("Default protocol should be https, got: %v", protocol)
	}

	if scm != "github" {
		t.Errorf("Default scm should be github, got: %v", scm)
	}

	if cloneType != "org" {
		t.Errorf("Default clone type should be org, got: %v", cloneType)
	}

	if namespace != "unset" {
		t.Errorf("Default gitlab namespace type should be unset, got: %v", namespace)
	}

}

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
