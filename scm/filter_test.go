package scm

import (
	"os"
	"testing"
)

func TestMatchingTopicsWithNoEnvTopics(t *testing.T) {
	os.Setenv("GHORG_TOPICS", "")

	t.Run("When repo topics are empty", func(tt *testing.T) {
		rpTopics := []string{}

		want := true
		got := hasMatchingTopic(rpTopics)
		if want != got {
			tt.Errorf("Expected %v repo, got: %v", want, got)
		}
	})

	t.Run("When any repo topics are set", func(tt *testing.T) {
		rpTopics := []string{"myTopic", "anotherTopic", "3rdTopic"}

		want := true
		got := hasMatchingTopic(rpTopics)
		if want != got {
			tt.Errorf("Expected %v repo, got: %v", want, got)
		}
	})
}

func TestMatchingTopicsWithSingleEnvTopic(t *testing.T) {
	os.Setenv("GHORG_TOPICS", "myTopic")

	t.Run("When repo topic is empty", func(tt *testing.T) {
		rpTopics := []string{}

		want := false
		got := hasMatchingTopic(rpTopics)
		if want != got {
			tt.Errorf("Expected %v repo, got: %v", want, got)
		}
	})

	t.Run("When single repo topic matches", func(tt *testing.T) {
		rpTopics := []string{"myTopic"}

		want := true
		got := hasMatchingTopic(rpTopics)
		if want != got {
			tt.Errorf("Expected %v repo, got: %v", want, got)
		}
	})

	t.Run("When one of multiple repo topics matches", func(tt *testing.T) {
		rpTopics := []string{"anotherTopic", "myTopic", "3rdTopic"}

		want := true
		got := hasMatchingTopic(rpTopics)
		if want != got {
			tt.Errorf("Expected %v repo, got: %v", want, got)
		}
	})

	os.Setenv("GHORG_TOPICS", "")
}

func TestMatchingTopicsWithMultipleEnvTopics(t *testing.T) {
	os.Setenv("GHORG_TOPICS", "myTopic,3rdTopic")

	t.Run("When repo topic is empty", func(tt *testing.T) {
		rpTopics := []string{}

		want := false
		got := hasMatchingTopic(rpTopics)
		if want != got {
			tt.Errorf("Expected %v repo, got: %v", want, got)
		}
	})

	t.Run("When Repo topic matches none", func(tt *testing.T) {
		rpTopics := []string{"anotherTopic"}

		want := false
		got := hasMatchingTopic(rpTopics)
		if want != got {
			tt.Errorf("Expected %v repo, got: %v", want, got)
		}
	})

	t.Run("When Repo topic matches at least one", func(tt *testing.T) {
		rpTopics := []string{"3rdTopic"}

		want := true
		got := hasMatchingTopic(rpTopics)
		if want != got {
			tt.Errorf("Expected %v repo, got: %v", want, got)
		}
	})

	t.Run("When Repo topic matches multiple", func(tt *testing.T) {
		rpTopics := []string{"3rdTopic", "myTopic"}

		want := true
		got := hasMatchingTopic(rpTopics)
		if want != got {
			tt.Errorf("Expected %v repo, got: %v", want, got)
		}
	})

	os.Setenv("GHORG_TOPICS", "")
}

func TestReplaceSSHHostname(t *testing.T) {
	t.Run("When newHostname is empty, return original URL unchanged", func(tt *testing.T) {
		originalURL := "git@github.com:org/repo.git"
		want := originalURL
		got := ReplaceSSHHostname(originalURL, "")
		if want != got {
			tt.Errorf("Expected %v, got: %v", want, got)
		}
	})

	t.Run("When replacing GitHub SSH hostname with colon separator", func(tt *testing.T) {
		originalURL := "git@github.com:org/repo.git"
		want := "git@my-github-alias:org/repo.git"
		got := ReplaceSSHHostname(originalURL, "my-github-alias")
		if want != got {
			tt.Errorf("Expected %v, got: %v", want, got)
		}
	})

	t.Run("When replacing GitLab SSH hostname", func(tt *testing.T) {
		originalURL := "git@gitlab.com:group/subgroup/repo.git"
		want := "git@custom.gitlab.host:group/subgroup/repo.git"
		got := ReplaceSSHHostname(originalURL, "custom.gitlab.host")
		if want != got {
			tt.Errorf("Expected %v, got: %v", want, got)
		}
	})

	t.Run("When replacing Bitbucket SSH hostname", func(tt *testing.T) {
		originalURL := "git@bitbucket.org:workspace/repo.git"
		want := "git@bitbucket-alias:workspace/repo.git"
		got := ReplaceSSHHostname(originalURL, "bitbucket-alias")
		if want != got {
			tt.Errorf("Expected %v, got: %v", want, got)
		}
	})

	t.Run("When replacing self-hosted GitLab SSH hostname", func(tt *testing.T) {
		originalURL := "git@gitlab.example.com:org/repo.git"
		want := "git@my-gitlab:org/repo.git"
		got := ReplaceSSHHostname(originalURL, "my-gitlab")
		if want != got {
			tt.Errorf("Expected %v, got: %v", want, got)
		}
	})

	t.Run("When URL has slash separator instead of colon", func(tt *testing.T) {
		originalURL := "git@git.sr.ht/~user/repo"
		want := "git@sourcehut-alias/~user/repo"
		got := ReplaceSSHHostname(originalURL, "sourcehut-alias")
		if want != got {
			tt.Errorf("Expected %v, got: %v", want, got)
		}
	})

	t.Run("When hostname has subdomain", func(tt *testing.T) {
		originalURL := "git@git.company.example.com:org/repo.git"
		want := "git@my-gitlab:org/repo.git"
		got := ReplaceSSHHostname(originalURL, "my-gitlab")
		if want != got {
			tt.Errorf("Expected %v, got: %v", want, got)
		}
	})
}
