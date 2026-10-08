package git

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/gabrie30/ghorg/scm"
)

// isolateGit keeps the developer's git config (commit signing, hooks, default branch name) out of these tests.
func isolateGit(t *testing.T) {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_AUTHOR_NAME", "ghorg test")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@ghorg.dev")
	t.Setenv("GIT_COMMITTER_NAME", "ghorg test")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@ghorg.dev")
	t.Setenv("GHORG_DEBUG", "")
}

// mustGit runs git in dir and fails the test if it errors.
func mustGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// newClonePair creates a seed repo with one commit on main, a bare origin cloned from it, and a working clone of
// origin. The seed lives next to work at filepath.Join(filepath.Dir(work), "seed").
func newClonePair(t *testing.T) (origin, work string) {
	t.Helper()
	isolateGit(t)
	base := t.TempDir()
	seed := filepath.Join(base, "seed")
	origin = filepath.Join(base, "origin.git")
	work = filepath.Join(base, "work")
	mustGit(t, base, "init", "-q", "-b", "main", seed)
	mustGit(t, seed, "commit", "-q", "--allow-empty", "-m", "initial")
	mustGit(t, base, "clone", "-q", "--bare", seed, origin)
	mustGit(t, base, "clone", "-q", origin, work)
	return origin, work
}

func TestCurrentBranchReportsBranchAndDetachedHead(t *testing.T) {
	_, work := newClonePair(t)
	g := NewGit()
	repo := scm.Repo{HostPath: work}

	branch, detached, err := g.CurrentBranch(repo)
	if err != nil || detached || branch != "main" {
		t.Fatalf("got (%q, %v, %v), want (\"main\", false, nil)", branch, detached, err)
	}

	mustGit(t, work, "checkout", "-q", "--detach")
	branch, detached, err = g.CurrentBranch(repo)
	if err != nil || !detached || branch != "" {
		t.Fatalf("got (%q, %v, %v), want (\"\", true, nil)", branch, detached, err)
	}
}

func TestCurrentBranchWorksBeforeFirstCommit(t *testing.T) {
	isolateGit(t)
	base := t.TempDir()
	dir := filepath.Join(base, "empty")
	mustGit(t, base, "init", "-q", "-b", "main", dir)
	g := NewGit()
	repo := scm.Repo{HostPath: dir}

	branch, detached, err := g.CurrentBranch(repo)
	if err != nil || detached || branch != "main" {
		t.Fatalf("got (%q, %v, %v), want (\"main\", false, nil)", branch, detached, err)
	}
	if g.LocalBranchExists(repo, "main") {
		t.Fatal("main has no commits yet so refs/heads/main should not exist")
	}
}

func TestUpstream(t *testing.T) {
	_, work := newClonePair(t)
	g := NewGit()
	repo := scm.Repo{HostPath: work}

	name, gone, err := g.Upstream(repo, "main")
	if err != nil || name != "origin/main" || gone {
		t.Fatalf("main: got (%q, %v, %v), want (\"origin/main\", false, nil)", name, gone, err)
	}

	mustGit(t, work, "branch", "-q", "no-upstream")
	name, gone, err = g.Upstream(repo, "no-upstream")
	if err != nil || name != "" || gone {
		t.Fatalf("no-upstream: got (%q, %v, %v), want (\"\", false, nil)", name, gone, err)
	}

	mustGit(t, work, "branch", "-q", "feat")
	mustGit(t, work, "config", "branch.feat.remote", "origin")
	mustGit(t, work, "config", "branch.feat.merge", "refs/heads/feat")
	name, gone, err = g.Upstream(repo, "feat")
	if err != nil || name != "origin/feat" || !gone {
		t.Fatalf("feat: got (%q, %v, %v), want (\"origin/feat\", true, nil)", name, gone, err)
	}
}

func TestAheadBehind(t *testing.T) {
	_, work := newClonePair(t)
	mustGit(t, work, "commit", "-q", "--allow-empty", "-m", "local")

	ahead, behind, err := NewGit().AheadBehind(scm.Repo{HostPath: work}, "main", "origin/main")
	if err != nil || ahead != 1 || behind != 0 {
		t.Fatalf("got (%d, %d, %v), want (1, 0, nil)", ahead, behind, err)
	}
}

func TestDefaultBranch(t *testing.T) {
	_, work := newClonePair(t)
	g := NewGit()
	repo := scm.Repo{HostPath: work}

	branch, err := g.DefaultBranch(repo)
	if err != nil || branch != "main" {
		t.Fatalf("got (%q, %v), want (\"main\", nil)", branch, err)
	}

	mustGit(t, work, "remote", "set-head", "origin", "-d")
	if _, err := g.DefaultBranch(repo); err == nil {
		t.Fatal("expected an error when origin/HEAD is not set")
	}
}

func TestBranchCheckedOutInWorktree(t *testing.T) {
	_, work := newClonePair(t)
	g := NewGit()
	repo := scm.Repo{HostPath: work}
	mustGit(t, work, "branch", "-q", "other")

	if checked, err := g.BranchCheckedOutInWorktree(repo, "main"); err != nil || !checked {
		t.Fatalf("main: got (%v, %v), want (true, nil)", checked, err)
	}
	if checked, err := g.BranchCheckedOutInWorktree(repo, "other"); err != nil || checked {
		t.Fatalf("other before worktree add: got (%v, %v), want (false, nil)", checked, err)
	}

	mustGit(t, work, "worktree", "add", "-q", filepath.Join(filepath.Dir(work), "wt"), "other")
	if checked, err := g.BranchCheckedOutInWorktree(repo, "other"); err != nil || !checked {
		t.Fatalf("other after worktree add: got (%v, %v), want (true, nil)", checked, err)
	}
}

func TestFastForwardBranchMovesBranchWithoutCheckout(t *testing.T) {
	origin, work := newClonePair(t)
	seed := filepath.Join(filepath.Dir(work), "seed")
	mustGit(t, seed, "commit", "-q", "--allow-empty", "-m", "second")
	mustGit(t, seed, "push", "-q", origin, "main")
	mustGit(t, work, "fetch", "-q", "origin")
	mustGit(t, work, "checkout", "-q", "-b", "other")

	if err := NewGit().FastForwardBranch(scm.Repo{HostPath: work}, "main"); err != nil {
		t.Fatalf("FastForwardBranch: %v", err)
	}
	if got, want := mustGit(t, work, "rev-parse", "main"), mustGit(t, work, "rev-parse", "origin/main"); got != want {
		t.Fatalf("main is at %s, want origin/main %s", got, want)
	}
	if got := mustGit(t, work, "symbolic-ref", "--short", "HEAD"); got != "other" {
		t.Fatalf("HEAD moved to %q, want other", got)
	}
}

func TestFetchOriginErrorIsFirstLineOfStderr(t *testing.T) {
	_, work := newClonePair(t)
	mustGit(t, work, "remote", "set-url", "origin", filepath.Join(filepath.Dir(work), "missing.git"))

	err := NewGit().FetchOrigin(scm.Repo{HostPath: work})
	if err == nil {
		t.Fatal("expected fetch from a missing origin to fail")
	}
	msg := err.Error()
	if strings.HasPrefix(msg, "fatal:") || strings.Contains(msg, "\n") || !strings.Contains(msg, "missing.git") {
		t.Fatalf("error should be one readable line naming missing.git, got %q", msg)
	}
}

func TestSSHBatchModeEnvWhenNothingConfigured(t *testing.T) {
	_, work := newClonePair(t)
	t.Setenv("GIT_SSH_COMMAND", "")
	t.Setenv("GIT_SSH", "")

	got := sshBatchModeEnv(scm.Repo{HostPath: work})

	if want := []string{"GIT_SSH_COMMAND=ssh -o BatchMode=yes"}; !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestSSHBatchModeEnvRespectsRepoSSHCommand(t *testing.T) {
	_, work := newClonePair(t)
	t.Setenv("GIT_SSH_COMMAND", "")
	t.Setenv("GIT_SSH", "")
	mustGit(t, work, "config", "core.sshCommand", "ssh -i ~/.ssh/work")

	if got := sshBatchModeEnv(scm.Repo{HostPath: work}); got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

func TestSSHBatchModeEnvRespectsEnv(t *testing.T) {
	_, work := newClonePair(t)
	t.Setenv("GIT_SSH_COMMAND", "ssh -i ~/.ssh/work")

	if got := sshBatchModeEnv(scm.Repo{HostPath: work}); got != nil {
		t.Fatalf("got %v, want nil", got)
	}
}

func TestFirstErrorLine(t *testing.T) {
	cases := []struct {
		name   string
		stderr string
		err    error
		want   string
	}{
		{"fatal prefix", "fatal: repository not found\n", errors.New("exit status 128"), "repository not found"},
		{"error prefix", "error: cannot lock ref\n", errors.New("exit status 1"), "cannot lock ref"},
		{"skips warnings and hints", "warning: redirecting\nhint: x\nfatal: Authentication failed for 'https://x/'\n", errors.New("exit status 128"), "Authentication failed for 'https://x/'"},
		{"ssh output", "git@github.com: Permission denied (publickey).\nfatal: Could not read from remote repository.\n", errors.New("exit status 128"), "git@github.com: Permission denied (publickey)."},
		{"skips fetch header", "From /tmp/x\n ! [rejected] main -> main (non-fast-forward)\n", errors.New("exit status 1"), "! [rejected] main -> main (non-fast-forward)"},
		{"empty falls back to error", "", errors.New("exit status 1"), "exit status 1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := firstErrorLine(tc.stderr, tc.err); got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDefaultBranchMissingTarget(t *testing.T) {
	_, work := newClonePair(t)
	mustGit(t, work, "update-ref", "-d", "refs/remotes/origin/main")
	if _, err := NewGit().DefaultBranch(scm.Repo{HostPath: work}); err == nil {
		t.Fatal("expected an error when origin/HEAD points at a missing branch")
	}
}

func TestCurrentBranchWithSameNamedTag(t *testing.T) {
	_, work := newClonePair(t)
	mustGit(t, work, "tag", "main")
	branch, detached, err := NewGit().CurrentBranch(scm.Repo{HostPath: work})
	if err != nil || detached || branch != "main" {
		t.Fatalf("got (%q, %v, %v), want (\"main\", false, nil)", branch, detached, err)
	}
}

func TestDiffHeadShowsStagedAndUnstagedChanges(t *testing.T) {
	_, work := newClonePair(t)
	t.Setenv("GHORG_COLOR", "")
	if err := os.WriteFile(filepath.Join(work, "a.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, work, "add", "a.txt")
	mustGit(t, work, "commit", "-q", "-m", "add a")
	if err := os.WriteFile(filepath.Join(work, "a.txt"), []byte("two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(work, "b.txt"), []byte("staged\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, work, "add", "b.txt")

	g := NewGit()
	repo := scm.Repo{HostPath: work}
	diff, err := g.DiffHead(repo)
	if err != nil {
		t.Fatalf("DiffHead: %v", err)
	}
	for _, want := range []string{"diff --git a/a.txt b/a.txt", "-one", "+two", "diff --git a/b.txt b/b.txt", "+staged"} {
		if !strings.Contains(diff, want) {
			t.Fatalf("diff is missing %q:\n%s", want, diff)
		}
	}
	if strings.Contains(diff, "\x1b[") {
		t.Fatal("diff has color codes even though GHORG_COLOR is not enabled")
	}

	t.Setenv("GHORG_COLOR", "enabled")
	prevNoColor := color.NoColor
	t.Cleanup(func() { color.NoColor = prevNoColor })
	color.NoColor = true
	plain, err := g.DiffHead(repo)
	if err != nil {
		t.Fatalf("DiffHead with NoColor: %v", err)
	}
	if strings.Contains(plain, "\x1b[") {
		t.Fatal("diff has color codes even though ghorg's color output is off")
	}

	color.NoColor = false
	colored, err := g.DiffHead(repo)
	if err != nil {
		t.Fatalf("DiffHead with color: %v", err)
	}
	if !strings.Contains(colored, "\x1b[") {
		t.Fatal("diff has no color codes even though GHORG_COLOR is enabled")
	}
}
