package configs_test

import (
	"testing"

	"github.com/gabrie30/ghorg/configs"
)

func TestDefaultSettings(t *testing.T) {
	config, err := configs.Load(nil)
	if err != nil {
		t.Fatal(err)
	}

	if config.CloneProtocol != "https" {
		t.Errorf("Default protocol should be https, got: %v", config.CloneProtocol)
	}

	if config.ScmType != "github" {
		t.Errorf("Default scm should be github, got: %v", config.ScmType)
	}

	if config.CloneType != "org" {
		t.Errorf("Default clone type should be org, got: %v", config.CloneType)
	}

}

func TestVerifyTokenSet(t *testing.T) {
	config := &configs.Config{
		ScmType:     "github",
		Token: "",
	}

	t.Run("When cloning github", func(tt *testing.T) {
		err := config.VerifyToken()
		if err != configs.ErrNoGitHubToken {
			tt.Errorf("Expected ErrNoGitHubTokenError, got: %v", err)
		}

	})

	config = &configs.Config{
		ScmType:     "gitlab",
		Token: "",
	}

	t.Run("When cloning gitlab", func(tt *testing.T) {
		err := config.VerifyToken()
		if err != configs.ErrNoGitLabToken {
			tt.Errorf("Expected ErrNoGitHubTokenError, got: %v", err)
		}

	})

	config = &configs.Config{
		ScmType:           "bitbucket",
		Token: "",
	}

	t.Run("When cloning bitbucket with no username", func(tt *testing.T) {
		err := config.VerifyToken()
		if err != configs.ErrNoBitbucketUsername {
			tt.Errorf("Expected ErrNoBitbucketUsername, got: %v", err)
		}

	})

	config = &configs.Config{
		ScmType:              "bitbucket",
		BitbucketUsername:    "bitbucketuser",
		Token: "",
	}

	t.Run("When cloning bitbucket with username but no app password", func(tt *testing.T) {
		err := config.VerifyToken()
		if err != configs.ErrNoBitbucketAppPassword {
			tt.Errorf("Expected ErrNoBitbucketAppPassword, got: %v", err)
		}

	})
}

func TestVerifyConfigsSetCorrectly(t *testing.T) {
	config := &configs.Config{
		CloneType:     "bot",
		CloneProtocol: "ssh",
		ScmType:       "github",
	}

	t.Run("When unsupported clone type", func(tt *testing.T) {
		err := config.VerifyClone()
		if err != configs.ErrIncorrectCloneType {
			tt.Errorf("Expected ErrIncorrectCloneType, got: %v", err)
		}

	})

	config = &configs.Config{
		CloneType:     "org",
		CloneProtocol: "ftp",
		ScmType:       "githubz",
	}

	t.Run("When unsupported protocol", func(tt *testing.T) {
		err := config.VerifyClone()
		if err != configs.ErrIncorrectProtocolType {
			tt.Errorf("Expected ErrIncorrectProtocolType, got: %v", err)
		}

	})
}
