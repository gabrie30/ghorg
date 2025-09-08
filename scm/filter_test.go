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
