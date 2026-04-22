package scm

import (
	"os"
	"testing"
)

func TestGithubListWorkerCount(t *testing.T) {
	t.Run("zero extra pages", func(t *testing.T) {
		os.Unsetenv("GHORG_GITHUB_REPO_LIST_CONCURRENCY")
		if w := githubListWorkerCount(0); w != 1 {
			t.Fatalf("want 1, got %d", w)
		}
	})
	t.Run("unlimited when env unset", func(t *testing.T) {
		os.Unsetenv("GHORG_GITHUB_REPO_LIST_CONCURRENCY")
		if w := githubListWorkerCount(100); w != 100 {
			t.Fatalf("want 100 (one worker per page), got %d", w)
		}
	})
	t.Run("honors explicit limit", func(t *testing.T) {
		os.Setenv("GHORG_GITHUB_REPO_LIST_CONCURRENCY", "3")
		t.Cleanup(func() { os.Unsetenv("GHORG_GITHUB_REPO_LIST_CONCURRENCY") })
		if w := githubListWorkerCount(10); w != 3 {
			t.Fatalf("want 3, got %d", w)
		}
		if w := githubListWorkerCount(1); w != 1 {
			t.Fatalf("want 1, got %d", w)
		}
	})
	t.Run("invalid env behaves like unlimited", func(t *testing.T) {
		os.Setenv("GHORG_GITHUB_REPO_LIST_CONCURRENCY", "nope")
		t.Cleanup(func() { os.Unsetenv("GHORG_GITHUB_REPO_LIST_CONCURRENCY") })
		if w := githubListWorkerCount(7); w != 7 {
			t.Fatalf("want 7, got %d", w)
		}
	})
	t.Run("caps explicit value", func(t *testing.T) {
		os.Setenv("GHORG_GITHUB_REPO_LIST_CONCURRENCY", "9999")
		t.Cleanup(func() { os.Unsetenv("GHORG_GITHUB_REPO_LIST_CONCURRENCY") })
		if w := githubListWorkerCount(10); w != 10 {
			t.Fatalf("want min(capped, extraPages)=10, got %d", w)
		}
		if w := githubListWorkerCount(800); w != maxGithubRepoListConcurrency {
			t.Fatalf("want %d, got %d", maxGithubRepoListConcurrency, w)
		}
	})
}
