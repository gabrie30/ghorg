package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/gabrie30/ghorg/git"
	"github.com/gabrie30/ghorg/scm"
)

// isolateGitConfig keeps the developer's git config (commit signing, hooks, default branch name) out of tests.
func isolateGitConfig(t *testing.T) {
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

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// pullFixture is a bare origin repo, a working clone that pull updates, and an upstream clone that pushes new
// commits to origin the way a teammate would.
type pullFixture struct {
	t        *testing.T
	base     string
	origin   string
	upstream string
	work     string
}

// newPullFixture creates origin with one commit on main containing README.md and a .gitignore that ignores
// ignored.txt, then clones it into work.
func newPullFixture(t *testing.T) *pullFixture {
	t.Helper()
	isolateGitConfig(t)
	base := t.TempDir()
	f := &pullFixture{
		t:        t,
		base:     base,
		origin:   filepath.Join(base, "origin.git"),
		upstream: filepath.Join(base, "upstream"),
		work:     filepath.Join(base, "work"),
	}
	mustGit(t, base, "init", "-q", "-b", "main", f.upstream)
	writeTestFile(t, filepath.Join(f.upstream, "README.md"), "hello\n")
	writeTestFile(t, filepath.Join(f.upstream, ".gitignore"), "ignored.txt\n")
	mustGit(t, f.upstream, "add", ".")
	mustGit(t, f.upstream, "commit", "-q", "-m", "initial")
	mustGit(t, base, "clone", "-q", "--bare", f.upstream, f.origin)
	mustGit(t, f.upstream, "remote", "add", "origin", f.origin)
	mustGit(t, f.upstream, "fetch", "-q", "origin")
	mustGit(t, base, "clone", "-q", f.origin, f.work)
	return f
}

func (f *pullFixture) branchExists(dir, branch string) bool {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	cmd.Dir = dir
	return cmd.Run() == nil
}

// pushUpstream commits a new file on branch in the upstream clone and pushes it to origin. The branch is created
// from main the first time.
func (f *pullFixture) pushUpstream(branch, file string) {
	f.t.Helper()
	if !f.branchExists(f.upstream, branch) {
		mustGit(f.t, f.upstream, "branch", branch, "main")
	}
	mustGit(f.t, f.upstream, "checkout", "-q", branch)
	writeTestFile(f.t, filepath.Join(f.upstream, file), file+"\n")
	mustGit(f.t, f.upstream, "add", file)
	mustGit(f.t, f.upstream, "commit", "-q", "-m", "add "+file)
	mustGit(f.t, f.upstream, "push", "-q", "origin", branch)
}

// commitWork commits a new file on the checked out branch of the working clone without pushing it.
func (f *pullFixture) commitWork(file string) {
	f.t.Helper()
	writeTestFile(f.t, filepath.Join(f.work, file), file+"\n")
	mustGit(f.t, f.work, "add", file)
	mustGit(f.t, f.work, "commit", "-q", "-m", "local "+file)
}

// checkoutTracking fetches and checks out branch in the working clone, tracking origin/<branch>.
func (f *pullFixture) checkoutTracking(branch string) {
	f.t.Helper()
	mustGit(f.t, f.work, "fetch", "-q", "origin")
	mustGit(f.t, f.work, "checkout", "-q", "-b", branch, "--track", "origin/"+branch)
}

func (f *pullFixture) workRef(ref string) string {
	f.t.Helper()
	return mustGit(f.t, f.work, "rev-parse", ref)
}

func (f *pullFixture) originRef(branch string) string {
	f.t.Helper()
	return mustGit(f.t, f.origin, "rev-parse", "refs/heads/"+branch)
}

// workState captures what a skip must not change: working tree status, every local branch, and HEAD.
func (f *pullFixture) workState() string {
	f.t.Helper()
	return strings.Join([]string{
		mustGit(f.t, f.work, "status", "--porcelain"),
		mustGit(f.t, f.work, "for-each-ref", "--format=%(refname) %(objectname)", "refs/heads"),
		mustGit(f.t, f.work, "rev-parse", "HEAD"),
	}, "\n")
}

func (f *pullFixture) pull() pullResult {
	f.t.Helper()
	return pullRepo(git.NewGit(), f.work, "work", false)
}

func (f *pullFixture) forcePull() pullResult {
	f.t.Helper()
	return pullRepo(git.NewGit(), f.work, "work", true)
}

func assertResult(t *testing.T, got pullResult, wantStatus pullStatus, wantNotes ...string) {
	t.Helper()
	if got.Status != wantStatus || !slices.Equal(got.Notes, wantNotes) {
		t.Fatalf("got status %d notes %q, want status %d notes %q", got.Status, got.Notes, wantStatus, wantNotes)
	}
}

func TestPullFastForwardsCurrentBranch(t *testing.T) {
	f := newPullFixture(t)
	f.pushUpstream("main", "a.txt")

	assertResult(t, f.pull(), pullUpdated, "main +1")
	if f.workRef("main") != f.originRef("main") {
		t.Fatal("main was not fast-forwarded to origin")
	}
}

func TestPullUpToDate(t *testing.T) {
	f := newPullFixture(t)
	assertResult(t, f.pull(), pullUpToDate)
}

func TestPullReportsUnpushedCommits(t *testing.T) {
	f := newPullFixture(t)
	f.commitWork("local.txt")
	assertResult(t, f.pull(), pullUpToDate, "main: 1 unpushed commits")
}

func TestPullSkipsUncommittedChanges(t *testing.T) {
	f := newPullFixture(t)
	f.pushUpstream("main", "a.txt")
	before := f.workRef("main")
	writeTestFile(t, filepath.Join(f.work, "README.md"), "local edit\n")

	assertResult(t, f.pull(), pullSkipped, "main: uncommitted changes (1 files)")
	if f.workRef("main") != before {
		t.Fatal("main moved even though the working tree was dirty")
	}
	if got := readTestFile(t, filepath.Join(f.work, "README.md")); got != "local edit\n" {
		t.Fatalf("local edit was changed to %q", got)
	}
}

func TestPullSkipsUntrackedFiles(t *testing.T) {
	f := newPullFixture(t)
	f.pushUpstream("main", "a.txt")
	writeTestFile(t, filepath.Join(f.work, "scratch.txt"), "mine\n")
	before := f.workState()

	assertResult(t, f.pull(), pullSkipped, "main: uncommitted changes (1 files)")
	if f.workState() != before {
		t.Fatal("skip changed the repo state")
	}
	if got := readTestFile(t, filepath.Join(f.work, "scratch.txt")); got != "mine\n" {
		t.Fatalf("untracked file was changed to %q", got)
	}
}

func TestPullSkipsBranchWithoutUpstreamButUpdatesDefaultBranch(t *testing.T) {
	f := newPullFixture(t)
	f.pushUpstream("main", "a.txt")
	mustGit(t, f.work, "checkout", "-q", "-b", "local")

	assertResult(t, f.pull(), pullSkipped, "local: no upstream branch", "main +1")
	if f.workRef("main") != f.originRef("main") {
		t.Fatal("main was not fast-forwarded to origin")
	}
}

func TestPullSkipsGoneUpstream(t *testing.T) {
	f := newPullFixture(t)
	f.pushUpstream("feat", "f.txt")
	f.checkoutTracking("feat")
	mustGit(t, f.upstream, "push", "-q", "origin", "--delete", "feat")
	mustGit(t, f.work, "update-ref", "-d", "refs/remotes/origin/feat")
	before := f.workState()

	assertResult(t, f.pull(), pullSkipped, "feat: upstream origin/feat no longer exists")
	if f.workState() != before {
		t.Fatal("skip changed the repo state")
	}
}

func TestPullSkipsDivergedBranch(t *testing.T) {
	f := newPullFixture(t)
	f.pushUpstream("main", "a.txt")
	f.commitWork("local.txt")
	before := f.workRef("main")

	assertResult(t, f.pull(), pullSkipped, "main: diverged from origin/main (1 ahead, 1 behind)")
	if f.workRef("main") != before {
		t.Fatal("diverged main was moved")
	}
}

func TestPullSkipsDetachedHeadButUpdatesDefaultBranch(t *testing.T) {
	f := newPullFixture(t)
	f.pushUpstream("main", "a.txt")
	mustGit(t, f.work, "checkout", "-q", "--detach")
	head := f.workRef("HEAD")

	assertResult(t, f.pull(), pullSkipped, "detached HEAD", "main +1")
	if f.workRef("HEAD") != head {
		t.Fatal("detached HEAD moved")
	}
	if f.workRef("main") != f.originRef("main") {
		t.Fatal("main was not fast-forwarded to origin")
	}
}

func TestPullSkipsUnfinishedOperations(t *testing.T) {
	cases := []struct {
		marker string
		isDir  bool
		want   string
	}{
		{"rebase-merge", true, "rebase in progress"},
		{"rebase-apply", true, "rebase in progress"},
		{"MERGE_HEAD", false, "merge in progress"},
		{"CHERRY_PICK_HEAD", false, "cherry-pick in progress"},
		{"REVERT_HEAD", false, "revert in progress"},
	}
	for _, tc := range cases {
		t.Run(tc.marker, func(t *testing.T) {
			f := newPullFixture(t)
			f.pushUpstream("main", "a.txt")
			before := f.workRef("main")
			path := filepath.Join(f.work, ".git", tc.marker)
			if tc.isDir {
				if err := os.MkdirAll(path, 0o755); err != nil {
					t.Fatal(err)
				}
			} else {
				writeTestFile(t, path, before+"\n")
			}

			assertResult(t, f.pull(), pullSkipped, tc.want)
			if f.workRef("main") != before {
				t.Fatal("main moved during an unfinished operation")
			}
		})
	}
}

func TestPullSkipsRepoWithoutOrigin(t *testing.T) {
	f := newPullFixture(t)
	mustGit(t, f.work, "remote", "rename", "origin", "upstream")
	before := f.workState()

	assertResult(t, f.pull(), pullSkipped, "no origin remote")
	if f.workState() != before {
		t.Fatal("skip changed the repo state")
	}
}

func TestPullFailsWhenFetchFails(t *testing.T) {
	f := newPullFixture(t)
	mustGit(t, f.work, "remote", "set-url", "origin", filepath.Join(f.base, "missing.git"))

	res := f.pull()
	if res.Status != pullFailed || len(res.Notes) != 1 {
		t.Fatalf("got status %d notes %q, want one failed note", res.Status, res.Notes)
	}
	note := res.Notes[0]
	if !strings.HasPrefix(note, "fetch failed: ") || strings.Contains(note, "fatal:") || !strings.Contains(note, "missing.git") {
		t.Fatalf("fetch failure note is not readable: %q", note)
	}
}

func TestPullSkipsEmptyRepo(t *testing.T) {
	f := newPullFixture(t)
	empty := filepath.Join(f.base, "empty")
	mustGit(t, f.base, "init", "-q", "-b", "main", empty)
	mustGit(t, empty, "remote", "add", "origin", f.origin)

	res := pullRepo(git.NewGit(), empty, "empty", false)
	if res.Status != pullSkipped || len(res.Notes) == 0 || res.Notes[0] != "main: no commits yet" {
		t.Fatalf("got status %d notes %q, want skipped starting with \"main: no commits yet\"", res.Status, res.Notes)
	}
	if f.branchExists(empty, "main") {
		t.Fatal("pull must not create branches")
	}
}

func TestPullFastForwardsDefaultBranchWhileCurrentBranchIsDirty(t *testing.T) {
	f := newPullFixture(t)
	f.pushUpstream("feat", "f.txt")
	f.checkoutTracking("feat")
	f.pushUpstream("main", "a.txt")
	writeTestFile(t, filepath.Join(f.work, "f.txt"), "local edit\n")

	assertResult(t, f.pull(), pullSkipped, "feat: uncommitted changes (1 files)", "main +1")
	if f.workRef("main") != f.originRef("main") {
		t.Fatal("main was not fast-forwarded to origin")
	}
	if got := mustGit(t, f.work, "symbolic-ref", "--short", "HEAD"); got != "feat" {
		t.Fatalf("checked out branch changed to %q", got)
	}
	if got := readTestFile(t, filepath.Join(f.work, "f.txt")); got != "local edit\n" {
		t.Fatalf("local edit was changed to %q", got)
	}
}

func TestPullSkipsDivergedDefaultBranch(t *testing.T) {
	f := newPullFixture(t)
	f.pushUpstream("feat", "f.txt")
	f.checkoutTracking("feat")
	mustGit(t, f.work, "checkout", "-q", "main")
	f.commitWork("local.txt")
	mustGit(t, f.work, "checkout", "-q", "feat")
	f.pushUpstream("main", "a.txt")
	before := f.workRef("main")

	assertResult(t, f.pull(), pullSkipped, "main: diverged from origin/main (1 ahead, 1 behind)")
	if f.workRef("main") != before {
		t.Fatal("diverged main was moved")
	}
}

func TestPullSkipsDefaultBranchCheckedOutInAnotherWorktree(t *testing.T) {
	f := newPullFixture(t)
	f.pushUpstream("feat", "f.txt")
	f.checkoutTracking("feat")
	mustGit(t, f.work, "worktree", "add", "-q", filepath.Join(f.base, "wt"), "main")
	f.pushUpstream("main", "a.txt")
	before := f.workRef("main")

	assertResult(t, f.pull(), pullSkipped, "main: checked out in another worktree")
	if f.workRef("main") != before {
		t.Fatal("main moved while checked out in another worktree")
	}
}

func TestPullDoesNotCreateMissingDefaultBranch(t *testing.T) {
	f := newPullFixture(t)
	f.pushUpstream("feat", "f.txt")
	f.checkoutTracking("feat")
	mustGit(t, f.work, "branch", "-q", "-D", "main")
	f.pushUpstream("main", "a.txt")

	assertResult(t, f.pull(), pullUpToDate)
	if f.branchExists(f.work, "main") {
		t.Fatal("pull created a local main branch")
	}
}

func TestPullSkipsWhenOriginHeadUnset(t *testing.T) {
	f := newPullFixture(t)
	// Newer git recreates origin/HEAD on fetch unless told not to.
	mustGit(t, f.work, "config", "remote.origin.followRemoteHEAD", "never")
	mustGit(t, f.work, "remote", "set-head", "origin", "-d")
	before := f.workState()

	assertResult(t, f.pull(), pullSkipped, noDefaultBranchReason)
	if f.workState() != before {
		t.Fatal("skip changed the repo state")
	}
}

func TestPullWorksWithGhorgDebug(t *testing.T) {
	f := newPullFixture(t)
	f.pushUpstream("main", "a.txt")
	t.Setenv("GHORG_DEBUG", "1")

	assertResult(t, f.pull(), pullUpdated, "main +1")
}

func TestForcePullSkipsWhenOriginHeadUnset(t *testing.T) {
	f := newPullFixture(t)
	mustGit(t, f.work, "config", "remote.origin.followRemoteHEAD", "never")
	mustGit(t, f.work, "remote", "set-head", "origin", "-d")
	writeTestFile(t, filepath.Join(f.work, "README.md"), "local edit\n")

	assertResult(t, f.forcePull(), pullSkipped, noDefaultBranchReason)
	if got := readTestFile(t, filepath.Join(f.work, "README.md")); got != "local edit\n" {
		t.Fatalf("local edit was changed to %q", got)
	}
}

func TestForcePullSkipsWhenDefaultBranchInAnotherWorktree(t *testing.T) {
	f := newPullFixture(t)
	f.pushUpstream("feat", "f.txt")
	f.checkoutTracking("feat")
	mustGit(t, f.work, "worktree", "add", "-q", filepath.Join(f.base, "wt"), "main")
	writeTestFile(t, filepath.Join(f.work, "README.md"), "local edit\n")

	assertResult(t, f.forcePull(), pullSkipped, "main: checked out in another worktree")
	if got := readTestFile(t, filepath.Join(f.work, "README.md")); got != "local edit\n" {
		t.Fatalf("local edit was changed to %q", got)
	}
}

func TestForcePullResetsToDefaultBranch(t *testing.T) {
	f := newPullFixture(t)
	f.pushUpstream("feat", "f.txt")
	f.checkoutTracking("feat")
	mustGit(t, f.work, "checkout", "-q", "main")
	f.commitWork("unpushed-main.txt")
	mustGit(t, f.work, "checkout", "-q", "feat")
	f.commitWork("unpushed-feat.txt")
	featBefore := f.workRef("feat")
	f.pushUpstream("main", "a.txt")
	writeTestFile(t, filepath.Join(f.work, "README.md"), "local edit\n")
	writeTestFile(t, filepath.Join(f.work, "scratch.txt"), "untracked\n")
	writeTestFile(t, filepath.Join(f.work, "ignored.txt"), "ignored\n")

	assertResult(t, f.forcePull(), pullUpdated, "reset to origin/main")

	if got := mustGit(t, f.work, "symbolic-ref", "--short", "HEAD"); got != "main" {
		t.Fatalf("checked out branch is %q, want main", got)
	}
	if f.workRef("main") != f.originRef("main") {
		t.Fatal("main does not match origin")
	}
	if f.workRef("feat") != featBefore {
		t.Fatal("feat was modified")
	}
	if got := readTestFile(t, filepath.Join(f.work, "README.md")); got != "hello\n" {
		t.Fatalf("README.md is %q, want the committed content", got)
	}
	if _, err := os.Stat(filepath.Join(f.work, "scratch.txt")); !os.IsNotExist(err) {
		t.Fatal("untracked file was not removed")
	}
	if got := readTestFile(t, filepath.Join(f.work, "ignored.txt")); got != "ignored\n" {
		t.Fatalf("ignored file changed to %q", got)
	}
}

func TestForcePullSkipsUnfinishedRebase(t *testing.T) {
	f := newPullFixture(t)
	writeTestFile(t, filepath.Join(f.work, "README.md"), "local edit\n")
	if err := os.MkdirAll(filepath.Join(f.work, ".git", "rebase-merge"), 0o755); err != nil {
		t.Fatal(err)
	}

	assertResult(t, f.forcePull(), pullSkipped, "rebase in progress")
	if got := readTestFile(t, filepath.Join(f.work, "README.md")); got != "local edit\n" {
		t.Fatalf("local edit was changed to %q", got)
	}
}

func TestForcePullLeavesRepoUntouchedWhenFetchFails(t *testing.T) {
	f := newPullFixture(t)
	writeTestFile(t, filepath.Join(f.work, "README.md"), "local edit\n")
	mustGit(t, f.work, "remote", "set-url", "origin", filepath.Join(f.base, "missing.git"))

	res := f.forcePull()
	if res.Status != pullFailed {
		t.Fatalf("got status %d notes %q, want failed", res.Status, res.Notes)
	}
	if got := readTestFile(t, filepath.Join(f.work, "README.md")); got != "local edit\n" {
		t.Fatalf("local edit was changed to %q after a failed fetch", got)
	}
}

// mkdirs creates each slash-separated path under root.
func mkdirs(t *testing.T, root string, rels ...string) {
	t.Helper()
	for _, rel := range rels {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(rel)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func TestDiscoverReposFindsWorkTreesAndBareRepos(t *testing.T) {
	root := t.TempDir()
	mkdirs(t, root,
		"api/.git",
		"api/vendor/nested/.git",
		"group/web/.git",
		"mirror.git/objects",
		"mirror.git/refs",
		"notes",
	)
	writeTestFile(t, filepath.Join(root, "mirror.git", "HEAD"), "ref: refs/heads/main\n")
	writeTestFile(t, filepath.Join(root, "linked", ".git"), "gitdir: /elsewhere/.git/worktrees/linked\n")

	repos, bare, err := discoverRepos(root)
	if err != nil {
		t.Fatal(err)
	}
	wantRepos := []string{
		filepath.Join(root, "api"),
		filepath.Join(root, "group", "web"),
		filepath.Join(root, "linked"),
	}
	if !slices.Equal(repos, wantRepos) {
		t.Fatalf("repos = %q, want %q", repos, wantRepos)
	}
	if wantBare := []string{filepath.Join(root, "mirror.git")}; !slices.Equal(bare, wantBare) {
		t.Fatalf("bare = %q, want %q", bare, wantBare)
	}
}

func TestDiscoverReposWhenRootIsARepo(t *testing.T) {
	root := t.TempDir()
	mkdirs(t, root, ".git", "sub/.git")

	repos, bare, err := discoverRepos(root)
	if err != nil || !slices.Equal(repos, []string{root}) || len(bare) != 0 {
		t.Fatalf("got (%q, %q, %v), want only the root", repos, bare, err)
	}
}

func TestDiscoverReposDoesNotFollowSymlinks(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	mkdirs(t, outside, "elsewhere/.git")
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skipf("symlinks not supported here: %v", err)
	}

	repos, bare, err := discoverRepos(root)
	if err != nil || len(repos) != 0 || len(bare) != 0 {
		t.Fatalf("got (%q, %q, %v), want nothing", repos, bare, err)
	}
}

func TestPullDisplayPath(t *testing.T) {
	root := filepath.Join("home", "me", "my-org")
	if got := pullDisplayPath(root, root); got != "my-org" {
		t.Fatalf("root repo: got %q, want my-org", got)
	}
	if got, want := pullDisplayPath(root, filepath.Join(root, "group", "web")), filepath.Join("group", "web"); got != want {
		t.Fatalf("nested repo: got %q, want %q", got, want)
	}
}

func TestPullProgressLine(t *testing.T) {
	cases := []struct {
		in   pullResult
		want string
	}{
		{pullResult{Path: "web", Status: pullUpdated, Notes: []string{"feature/login +1", "main +7"}}, "Success pulling web: feature/login +1; main +7"},
		{pullResult{Path: "api", Status: pullUpToDate}, "Up to date api"},
		{pullResult{Path: "api", Status: pullUpToDate, Notes: []string{"main: 2 unpushed commits"}}, "Up to date api: main: 2 unpushed commits"},
		{pullResult{Path: "infra", Status: pullSkipped, Notes: []string{"detached HEAD"}}, "Skipped infra: detached HEAD"},
		{pullResult{Path: "legacy", Status: pullFailed, Notes: []string{"fetch failed: boom"}}, "Failed legacy: fetch failed: boom"},
	}
	for _, tc := range cases {
		if got := pullProgressLine(tc.in); got != tc.want {
			t.Errorf("pullProgressLine(%+v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestPullSummary(t *testing.T) {
	results := []pullResult{
		{Path: "web", Status: pullUpdated, Notes: []string{"main +7"}},
		{Path: "legacy", Status: pullFailed, Notes: []string{"fetch failed: boom"}},
		{Path: "infra", Status: pullSkipped, Notes: []string{"main: diverged from origin/main (1 ahead, 3 behind)"}},
		{Path: "api", Status: pullUpToDate},
		{Path: "billing", Status: pullSkipped, Notes: []string{"feature/x: uncommitted changes (4 files)", "main +2"}},
	}
	want := "\nSkipped (2)\n" +
		"  billing  feature/x: uncommitted changes (4 files); main +2\n" +
		"  infra    main: diverged from origin/main (1 ahead, 3 behind)\n" +
		"\nFailed (1)\n" +
		"  legacy  fetch failed: boom\n" +
		"\nUpdated 1, up to date 1, skipped 2, failed 1"
	if got := pullSummary(results); got != want {
		t.Fatalf("pullSummary mismatch\n got: %q\nwant: %q", got, want)
	}
}

func TestPullSummaryWithNothingToReport(t *testing.T) {
	results := []pullResult{{Path: "api", Status: pullUpToDate}, {Path: "web", Status: pullUpdated, Notes: []string{"main +1"}}}
	if got, want := pullSummary(results), "\nUpdated 1, up to date 1, skipped 0, failed 0"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestPullExitCode(t *testing.T) {
	skippedOnly := []pullResult{{Status: pullUpdated}, {Status: pullSkipped}, {Status: pullUpToDate}}
	if got := pullExitCode(skippedOnly); got != 0 {
		t.Fatalf("skips only: got %d, want 0", got)
	}
	withFailure := append(skippedOnly, pullResult{Status: pullFailed})
	if got := pullExitCode(withFailure); got != 1 {
		t.Fatalf("with a failure: got %d, want 1", got)
	}
}

func TestRunPullsUpdatesEveryRepoUnderRoot(t *testing.T) {
	f := newPullFixture(t)
	root := t.TempDir()
	mustGit(t, root, "clone", "-q", f.origin, filepath.Join(root, "api"))
	mustGit(t, root, "clone", "-q", f.origin, filepath.Join(root, "group", "web"))
	f.pushUpstream("main", "a.txt")

	repos, bare, err := discoverRepos(root)
	if err != nil || len(repos) != 2 || len(bare) != 0 {
		t.Fatalf("discoverRepos = (%q, %q, %v), want two repos", repos, bare, err)
	}

	results := runPulls(git.NewGit(), root, repos, pullOptions{}, 2)
	wantPaths := []string{"api", filepath.Join("group", "web")}
	if len(results) != len(wantPaths) {
		t.Fatalf("got %d results, want %d", len(results), len(wantPaths))
	}
	for i, want := range wantPaths {
		if results[i].Path != want {
			t.Fatalf("results[%d].Path = %q, want %q", i, results[i].Path, want)
		}
		assertResult(t, results[i], pullUpdated, "main +1")
	}
}

func TestResolvePullRoot(t *testing.T) {
	dir := t.TempDir()
	want, err := filepath.EvalSymlinks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := resolvePullRoot(dir); err != nil || got != want {
		t.Fatalf("existing dir: got (%q, %v), want (%q, nil)", got, err, want)
	}
	if _, err := resolvePullRoot(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("missing dir: expected an error")
	}
	file := filepath.Join(dir, "file.txt")
	writeTestFile(t, file, "x")
	if _, err := resolvePullRoot(file); err == nil {
		t.Fatal("file: expected an error")
	}
}

func TestDisablePullPrompts(t *testing.T) {
	t.Setenv("GIT_TERMINAL_PROMPT", "")
	t.Setenv("GCM_INTERACTIVE", "")
	t.Setenv("GIT_SSH_COMMAND", "")

	disablePullPrompts()

	if got := os.Getenv("GCM_INTERACTIVE"); got != "never" {
		t.Fatalf("GCM_INTERACTIVE = %q, want never", got)
	}

	if got := os.Getenv("GIT_TERMINAL_PROMPT"); got != "0" {
		t.Fatalf("GIT_TERMINAL_PROMPT = %q, want 0", got)
	}
	if got := os.Getenv("GIT_SSH_COMMAND"); got != "" {
		t.Fatalf("GIT_SSH_COMMAND = %q, want it untouched", got)
	}
}

// newWorktreeRoot clones f.origin into root/app (and root/lib when withLib) and adds a worktree root/app-wt on a new
// branch called other.
func newWorktreeRoot(t *testing.T, f *pullFixture, withLib bool) string {
	t.Helper()
	root := t.TempDir()
	mustGit(t, root, "clone", "-q", f.origin, filepath.Join(root, "app"))
	if withLib {
		mustGit(t, root, "clone", "-q", f.origin, filepath.Join(root, "lib"))
	}
	mustGit(t, filepath.Join(root, "app"), "worktree", "add", "-q", "-b", "other", filepath.Join(root, "app-wt"))
	return root
}

func TestGroupBySharedGitDir(t *testing.T) {
	f := newPullFixture(t)
	root := newWorktreeRoot(t, f, true)

	repos, _, err := discoverRepos(root)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{filepath.Join(root, "app"), filepath.Join(root, "app-wt"), filepath.Join(root, "lib")}
	if !slices.Equal(repos, want) {
		t.Fatalf("repos = %q, want %q", repos, want)
	}
	got := groupBySharedGitDir(repos)
	if wantGroups := [][]int{{0, 1}, {2}}; !slices.EqualFunc(got, wantGroups, slices.Equal[[]int]) {
		t.Fatalf("groups = %v, want %v", got, wantGroups)
	}
}

func TestRunPullsSerializesWorktreesOfOneRepo(t *testing.T) {
	f := newPullFixture(t)
	root := newWorktreeRoot(t, f, false)
	f.pushUpstream("main", "a.txt")

	repos, _, err := discoverRepos(root)
	if err != nil || len(repos) != 2 {
		t.Fatalf("discoverRepos = (%q, %v), want two repos", repos, err)
	}
	results := runPulls(git.NewGit(), root, repos, pullOptions{}, 2)
	assertResult(t, results[0], pullUpdated, "main +1")
	if results[0].Path != "app" {
		t.Fatalf("results[0].Path = %q, want app", results[0].Path)
	}
	assertResult(t, results[1], pullSkipped, "other: no upstream branch")
	if results[1].Path != "app-wt" {
		t.Fatalf("results[1].Path = %q, want app-wt", results[1].Path)
	}
}

// failingDiffGitter is a real GitClient whose DiffHead always fails.
type failingDiffGitter struct{ git.GitClient }

func (failingDiffGitter) DiffHead(scm.Repo) (string, error) { return "", errors.New("boom") }

func TestPullMarksOnlyLocalChangeSkipsDirty(t *testing.T) {
	dirty := newPullFixture(t)
	writeTestFile(t, filepath.Join(dirty.work, "README.md"), "local edit\n")
	if !dirty.pull().Dirty {
		t.Fatal("a repo skipped for uncommitted changes must be marked Dirty")
	}

	diverged := newPullFixture(t)
	diverged.pushUpstream("main", "a.txt")
	diverged.commitWork("local.txt")
	if diverged.pull().Dirty {
		t.Fatal("a diverged but clean repo must not be marked Dirty")
	}
}

func TestLocalChangeLines(t *testing.T) {
	f := newPullFixture(t)
	t.Setenv("GHORG_COLOR", "")
	writeTestFile(t, filepath.Join(f.work, "README.md"), "local edit\n")
	writeTestFile(t, filepath.Join(f.work, "scratch.txt"), "mine\n")
	writeTestFile(t, filepath.Join(f.work, "build", "out.txt"), "artifact\n")

	lines := localChangeLines(git.NewGit(), f.work)
	if len(lines) == 0 || lines[0] != "untracked (2): build/, scratch.txt" {
		t.Fatalf("first line should list untracked files, got %q", lines)
	}
	for _, want := range []string{"diff --git a/README.md b/README.md", "-hello", "+local edit"} {
		if !slices.Contains(lines, want) {
			t.Fatalf("lines are missing %q: %q", want, lines)
		}
	}
}

func TestLocalChangeLinesWithoutUntrackedFiles(t *testing.T) {
	f := newPullFixture(t)
	t.Setenv("GHORG_COLOR", "")
	writeTestFile(t, filepath.Join(f.work, "README.md"), "local edit\n")

	lines := localChangeLines(git.NewGit(), f.work)
	if len(lines) == 0 || lines[0] != "diff --git a/README.md b/README.md" {
		t.Fatalf("with no untracked files the diff should come first, got %q", lines)
	}
}

func TestLocalChangeLinesTruncatesLongDiffs(t *testing.T) {
	f := newPullFixture(t)
	t.Setenv("GHORG_COLOR", "")
	var b strings.Builder
	for i := range 200 {
		fmt.Fprintf(&b, "line %d\n", i)
	}
	writeTestFile(t, filepath.Join(f.work, "README.md"), b.String())

	lines := localChangeLines(git.NewGit(), f.work)
	if len(lines) != maxDiffLines+1 {
		t.Fatalf("got %d lines, want %d diff lines plus the truncation hint", len(lines), maxDiffLines)
	}
	if lines[maxDiffLines-1] != "+line 93" {
		t.Fatalf("last kept diff line is %q, want %q", lines[maxDiffLines-1], "+line 93")
	}
	want := fmt.Sprintf("... diff truncated (106 more lines), run: git -C \"%s\" diff HEAD", f.work)
	if last := lines[len(lines)-1]; last != want {
		t.Fatalf("truncation hint is %q, want %q", last, want)
	}
}

func TestLocalChangeLinesReportsDiffError(t *testing.T) {
	f := newPullFixture(t)
	writeTestFile(t, filepath.Join(f.work, "README.md"), "local edit\n")

	lines := localChangeLines(failingDiffGitter{git.NewGit()}, f.work)
	if want := []string{"could not read diff: boom"}; !slices.Equal(lines, want) {
		t.Fatalf("got %q, want %q", lines, want)
	}
}

func TestRunPullsCollectsLocalChangesOnlyWhenVerbose(t *testing.T) {
	f := newPullFixture(t)
	t.Setenv("GHORG_COLOR", "")
	root := t.TempDir()
	app := filepath.Join(root, "app")
	mustGit(t, root, "clone", "-q", f.origin, app)
	writeTestFile(t, filepath.Join(app, "README.md"), "local edit\n")
	repos := []string{app}

	quiet := runPulls(git.NewGit(), root, repos, pullOptions{}, 1)
	if quiet[0].Status != pullSkipped || quiet[0].LocalChanges != nil {
		t.Fatalf("without verbose: got status %d local changes %q, want skipped and none", quiet[0].Status, quiet[0].LocalChanges)
	}

	verbose := runPulls(git.NewGit(), root, repos, pullOptions{verbose: true}, 1)
	if verbose[0].Status != pullSkipped || !slices.Contains(verbose[0].LocalChanges, "+local edit") {
		t.Fatalf("with verbose: got status %d local changes %q, want skipped with the diff", verbose[0].Status, verbose[0].LocalChanges)
	}
}

func TestPullSummaryPrintsLocalChanges(t *testing.T) {
	results := []pullResult{
		{Path: "web", Status: pullUpdated, Notes: []string{"main +1"}},
		{
			Path:         "billing",
			Status:       pullSkipped,
			Notes:        []string{"main: uncommitted changes (2 files)"},
			Dirty:        true,
			LocalChanges: []string{"untracked (1): notes.txt", "diff --git a/x b/x", "+new"},
		},
		{Path: "infra", Status: pullSkipped, Notes: []string{"detached HEAD"}},
	}
	want := "\nSkipped (2)\n" +
		"  billing  main: uncommitted changes (2 files)\n" +
		"    untracked (1): notes.txt\n" +
		"    diff --git a/x b/x\n" +
		"    +new\n" +
		"  infra    detached HEAD\n" +
		"\nUpdated 1, up to date 0, skipped 2, failed 0"
	if got := pullSummary(results); got != want {
		t.Fatalf("pullSummary mismatch\n got: %q\nwant: %q", got, want)
	}
}

func TestPullFlagShorthands(t *testing.T) {
	for short, long := range map[string]string{"f": "force", "v": "verbose"} {
		flag := pullCmd.Flags().ShorthandLookup(short)
		if flag == nil || flag.Name != long {
			t.Fatalf("-%s should be the shorthand for --%s, got %v", short, long, flag)
		}
	}
}
