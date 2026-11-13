package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gabrie30/ghorg/scm"
)

// commitFile commits a new file named file in dir.
func commitFile(t *testing.T, dir, file string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, file), []byte(file+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, dir, "add", file)
	mustGit(t, dir, "commit", "-q", "-m", "add "+file)
}

func TestHasLocalOnlyCommits(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(t *testing.T, work string)
		want  bool
	}{
		{
			name:  "fresh clone",
			setup: func(t *testing.T, work string) {},
			want:  false,
		},
		{
			name: "commit on another local branch",
			setup: func(t *testing.T, work string) {
				mustGit(t, work, "checkout", "-q", "-b", "feature")
				commitFile(t, work, "feature.txt")
				mustGit(t, work, "checkout", "-q", "main")
			},
			want: true,
		},
		{
			name: "commit on a detached HEAD",
			setup: func(t *testing.T, work string) {
				mustGit(t, work, "checkout", "-q", "--detach")
				commitFile(t, work, "detached.txt")
			},
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, work := newClonePair(t)
			tc.setup(t, work)

			got, err := NewGit().HasLocalOnlyCommits(scm.Repo{HostPath: work})
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("HasLocalOnlyCommits() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHasLocalOnlyCommitsBeforeFirstCommit(t *testing.T) {
	isolateGit(t)
	repo := filepath.Join(t.TempDir(), "repo")
	mustGit(t, filepath.Dir(repo), "init", "-q", "-b", "main", repo)

	got, err := NewGit().HasLocalOnlyCommits(scm.Repo{HostPath: repo})
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Error("HasLocalOnlyCommits() = true for a repo with no commits, want false")
	}
}

func TestHasStash(t *testing.T) {
	_, work := newClonePair(t)
	g := NewGit()

	got, err := g.HasStash(scm.Repo{HostPath: work})
	if err != nil {
		t.Fatal(err)
	}
	if got {
		t.Error("HasStash() = true for a fresh clone, want false")
	}

	if err := os.WriteFile(filepath.Join(work, "wip.txt"), []byte("wip\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, work, "stash", "-q", "--include-untracked")

	got, err = g.HasStash(scm.Repo{HostPath: work})
	if err != nil {
		t.Fatal(err)
	}
	if !got {
		t.Error("HasStash() = false after stashing changes, want true")
	}
}
