package configs

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"testing"
)

func createTestConf() string {
	os.Setenv("GHORG_ENV", "test")

	dir, err := ioutil.TempDir("", "ghorg_test")
	if err != nil {
		log.Fatal(err)
	}

	GhorgTestDir = dir

	srcFile := "../sample-conf.yaml"
	cpCmd := exec.Command("cp", srcFile, dir)

	mvCmd := exec.Command("mv", dir+"/sample-conf.yaml", dir+"/conf.yaml")
	err = cpCmd.Run()

	if err != nil {
		fmt.Println("could not copy sample-conf.yaml")
	}

	err = mvCmd.Run()

	if err != nil {
		fmt.Println("could not rename sample-conf.yaml")
	}

	initConfig()

	return dir
}

func TestDefaultSettings(t *testing.T) {
	defer os.RemoveAll(createTestConf())

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

		err := VerifyTokenSet()
		if err != ErrNoGitHubToken {
			tt.Errorf("Expected ErrNoGitHubTokenError, got: %v", err)
		}

	})

	t.Run("When cloning gitlab", func(tt *testing.T) {
		os.Setenv("GHORG_SCM_TYPE", "gitlab")
		os.Setenv("GHORG_GITLAB_TOKEN", "")

		err := VerifyTokenSet()
		if err != ErrNoGitLabToken {
			tt.Errorf("Expected ErrNoGitHubTokenError, got: %v", err)
		}

	})

	t.Run("When cloning bitbucket with no username", func(tt *testing.T) {
		os.Setenv("GHORG_SCM_TYPE", "bitbucket")
		os.Setenv("GHORG_BITBUCKET_USERNAME", "")
		err := VerifyTokenSet()
		if err != ErrNoBitbucketUsername {
			tt.Errorf("Expected ErrNoBitbucketUsername, got: %v", err)
		}

	})

	t.Run("When cloning bitbucket with username but no app password", func(tt *testing.T) {
		os.Setenv("GHORG_SCM_TYPE", "bitbucket")
		os.Setenv("GHORG_BITBUCKET_USERNAME", "bitbucketuser")
		os.Setenv("GHORG_BITBUCKET_APP_PASSWORD", "")
		err := VerifyTokenSet()
		if err != ErrNoBitbucketAppPassword {
			tt.Errorf("Expected ErrNoBitbucketAppPassword, got: %v", err)
		}

	})
}

func TestVerifyConfigsSetCorrectly(t *testing.T) {

	t.Run("When unsupported scm", func(tt *testing.T) {
		os.Setenv("GHORG_CLONE_TYPE", "org")
		os.Setenv("GHORG_CLONE_PROTOCOL", "ssh")

		os.Setenv("GHORG_SCM_TYPE", "githubz")

		err := VerifyConfigsSetCorrectly()
		if err != ErrIncorrectScmType {
			tt.Errorf("Expected ErrIncorrectScmType, got: %v", err)
		}

	})

	t.Run("When unsupported clone type", func(tt *testing.T) {
		os.Setenv("GHORG_SCM_TYPE", "github")
		os.Setenv("GHORG_CLONE_PROTOCOL", "ssh")

		os.Setenv("GHORG_CLONE_TYPE", "bot")

		err := VerifyConfigsSetCorrectly()
		if err != ErrIncorrectCloneType {
			tt.Errorf("Expected ErrIncorrectCloneType, got: %v", err)
		}

	})

	t.Run("When unsupported protocol", func(tt *testing.T) {
		os.Setenv("GHORG_SCM_TYPE", "github")
		os.Setenv("GHORG_CLONE_TYPE", "org")

		os.Setenv("GHORG_CLONE_PROTOCOL", "ftp")

		err := VerifyConfigsSetCorrectly()
		if err != ErrIncorrectProtocolType {
			tt.Errorf("Expected ErrIncorrectProtocolType, got: %v", err)
		}

	})
}
