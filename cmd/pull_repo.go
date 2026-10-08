package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabrie30/ghorg/scm"
)

// pullStatus is the overall outcome of pulling one repo. A higher value takes precedence over a lower one.
type pullStatus int

const (
	pullUpToDate pullStatus = iota
	pullUpdated
	pullSkipped
	pullFailed
)

// pullResult is the outcome of pulling one repo. Notes describe what happened to each branch, in order.
type pullResult struct {
	Path   string
	Status pullStatus
	Notes  []string
	// Dirty is set when the current branch was skipped for uncommitted changes or untracked files.
	Dirty bool
	// LocalChanges holds the lines --verbose prints under the repo in the summary.
	LocalChanges []string
}

func (r *pullResult) note(msg string) {
	r.Notes = append(r.Notes, msg)
}

func (r *pullResult) markUpdated(msg string) {
	r.Status = max(r.Status, pullUpdated)
	r.note(msg)
}

func (r *pullResult) markSkipped(msg string) {
	r.Status = max(r.Status, pullSkipped)
	r.note(msg)
}

func (r *pullResult) markFailed(msg string) {
	r.Status = pullFailed
	r.note(msg)
}

// pullGitter is the set of git operations ghorg pull uses. git.GitClient satisfies it.
type pullGitter interface {
	AbsoluteGitDir(scm.Repo) (string, error)
	HasRemote(scm.Repo, string) bool
	FetchOrigin(scm.Repo) error
	CurrentBranch(scm.Repo) (string, bool, error)
	LocalBranchExists(scm.Repo, string) bool
	WorkingTreeStatus(scm.Repo) (string, error)
	Upstream(scm.Repo, string) (string, bool, error)
	AheadBehind(scm.Repo, string, string) (int, int, error)
	MergeFastForwardOnly(scm.Repo, string) error
	DefaultBranch(scm.Repo) (string, error)
	BranchCheckedOutInWorktree(scm.Repo, string) (bool, error)
	FastForwardBranch(scm.Repo, string) error
	ResetHardHead(scm.Repo) error
	Clean(scm.Repo) error
	Checkout(scm.Repo) error
	Reset(scm.Repo) error
	DiffHead(scm.Repo) (string, error)
}

const noDefaultBranchReason = "cannot determine default branch (origin/HEAD not set or points to a missing branch, run: git remote set-head origin --auto)"

// inProgressMarkers are the paths git keeps in the git dir while an operation is unfinished.
var inProgressMarkers = []struct{ file, operation string }{
	{"rebase-merge", "rebase"},
	{"rebase-apply", "rebase"},
	{"MERGE_HEAD", "merge"},
	{"CHERRY_PICK_HEAD", "cherry-pick"},
	{"REVERT_HEAD", "revert"},
}

// pullRepo updates the repo at absPath. In safe mode it only fast-forwards and skips anything with local changes.
// With force it discards local changes and resets to origin's default branch.
// Both modes fetch first, so an auth or network failure never costs local changes. displayPath is how the repo is
// named in output.
func pullRepo(g pullGitter, absPath, displayPath string, force bool) pullResult {
	res := pullResult{Path: displayPath}
	repo := scm.Repo{Name: filepath.Base(absPath), HostPath: absPath}

	skipReason, err := pullPreflight(g, repo)
	if err != nil {
		res.markFailed(err.Error())
		return res
	}
	if skipReason != "" {
		res.markSkipped(skipReason)
		return res
	}

	if err := g.FetchOrigin(repo); err != nil {
		res.markFailed("fetch failed: " + err.Error())
		return res
	}

	if force {
		forcePull(g, repo, &res)
		return res
	}

	current := pullCurrentBranch(g, repo, &res)
	pullDefaultBranch(g, repo, current, &res)
	return res
}

// pullPreflight returns a reason to skip the repo when it has no origin or an operation such as a rebase is
// unfinished. An error means the directory could not be read as a git repo.
func pullPreflight(g pullGitter, repo scm.Repo) (string, error) {
	gitDir, err := g.AbsoluteGitDir(repo)
	if err != nil {
		return "", err
	}
	if !g.HasRemote(repo, "origin") {
		return "no origin remote", nil
	}
	for _, m := range inProgressMarkers {
		if _, err := os.Stat(filepath.Join(gitDir, m.file)); err == nil {
			return m.operation + " in progress", nil
		}
	}
	return "", nil
}

// pullCurrentBranch fast-forwards the checked out branch from its upstream when nothing local is at risk. It returns
// the branch name, or "" when HEAD is detached or the branch could not be read.
func pullCurrentBranch(g pullGitter, repo scm.Repo, res *pullResult) string {
	branch, detached, err := g.CurrentBranch(repo)
	if err != nil {
		res.markFailed("could not read current branch: " + err.Error())
		return ""
	}
	if detached {
		res.markSkipped("detached HEAD")
		return ""
	}
	if !g.LocalBranchExists(repo, branch) {
		res.markSkipped(branch + ": no commits yet")
		return branch
	}

	status, err := g.WorkingTreeStatus(repo)
	if err != nil {
		res.markFailed(fmt.Sprintf("%s: could not read status: %v", branch, err))
		return branch
	}
	if status != "" {
		res.Dirty = true
		res.markSkipped(fmt.Sprintf("%s: uncommitted changes (%d files)", branch, len(strings.Split(status, "\n"))))
		return branch
	}

	upstream, gone, err := g.Upstream(repo, branch)
	if err != nil {
		res.markFailed(fmt.Sprintf("%s: could not read upstream: %v", branch, err))
		return branch
	}
	if upstream == "" {
		res.markSkipped(branch + ": no upstream branch")
		return branch
	}
	if gone {
		res.markSkipped(fmt.Sprintf("%s: upstream %s no longer exists", branch, upstream))
		return branch
	}

	ahead, behind, err := g.AheadBehind(repo, "refs/heads/"+branch, upstream)
	if err != nil {
		res.markFailed(fmt.Sprintf("%s: could not compare with %s: %v", branch, upstream, err))
		return branch
	}
	switch {
	case ahead > 0 && behind > 0:
		res.markSkipped(fmt.Sprintf("%s: diverged from %s (%d ahead, %d behind)", branch, upstream, ahead, behind))
	case behind > 0:
		if err := g.MergeFastForwardOnly(repo, upstream); err != nil {
			res.markFailed(fmt.Sprintf("%s: fast-forward failed: %v", branch, err))
		} else {
			res.markUpdated(fmt.Sprintf("%s +%d", branch, behind))
		}
	case ahead > 0:
		res.note(fmt.Sprintf("%s: %d unpushed commits", branch, ahead))
	}
	return branch
}

// pullDefaultBranch fast-forwards the local default branch without checking it out, so it never touches files.
// current is the checked out branch, or "" when HEAD is detached.
func pullDefaultBranch(g pullGitter, repo scm.Repo, current string, res *pullResult) {
	def, err := g.DefaultBranch(repo)
	if err != nil {
		res.markSkipped(noDefaultBranchReason)
		return
	}
	if def == current || !g.LocalBranchExists(repo, def) {
		return
	}

	ahead, behind, err := g.AheadBehind(repo, "refs/heads/"+def, "refs/remotes/origin/"+def)
	if err != nil {
		res.markFailed(fmt.Sprintf("%s: could not compare with origin/%s: %v", def, def, err))
		return
	}
	if behind == 0 {
		return
	}
	if ahead > 0 {
		res.markSkipped(fmt.Sprintf("%s: diverged from origin/%s (%d ahead, %d behind)", def, def, ahead, behind))
		return
	}

	checkedOut, err := g.BranchCheckedOutInWorktree(repo, def)
	if err != nil {
		res.markFailed(fmt.Sprintf("%s: could not list worktrees: %v", def, err))
		return
	}
	if checkedOut {
		res.markSkipped(def + ": checked out in another worktree")
		return
	}

	if err := g.FastForwardBranch(repo, def); err != nil {
		res.markFailed(fmt.Sprintf("%s: fast-forward failed: %v", def, err))
		return
	}
	res.markUpdated(fmt.Sprintf("%s +%d", def, behind))
}

// forcePull discards local changes and leaves the repo on origin's default branch, much like ghorg clone updates an
// existing repo but without the final pull, since the fetch already ran. Reset and clean run before the checkout so
// local changes cannot block it. Other branches are not modified. It skips the repo when the default branch is
// checked out in another worktree, because the checkout would fail after local changes were already discarded.
func forcePull(g pullGitter, repo scm.Repo, res *pullResult) {
	def, err := g.DefaultBranch(repo)
	if err != nil {
		res.markSkipped(noDefaultBranchReason)
		return
	}

	current, _, err := g.CurrentBranch(repo)
	if err != nil {
		res.markFailed("could not read current branch: " + err.Error())
		return
	}
	if current != def {
		checkedOut, err := g.BranchCheckedOutInWorktree(repo, def)
		if err != nil {
			res.markFailed(fmt.Sprintf("%s: could not list worktrees: %v", def, err))
			return
		}
		if checkedOut {
			res.markSkipped(def + ": checked out in another worktree")
			return
		}
	}
	repo.CloneBranch = def

	steps := []struct {
		name string
		run  func(scm.Repo) error
	}{
		{"git reset --hard", g.ResetHardHead},
		{"git clean -f -d", g.Clean},
		{"git checkout " + def, g.Checkout},
		{"git reset --hard origin/" + def, g.Reset},
	}
	for _, step := range steps {
		if err := step.run(repo); err != nil {
			res.markFailed(fmt.Sprintf("%s failed: %v", step.name, err))
			return
		}
	}
	res.markUpdated("reset to origin/" + def)
}

// maxDiffLines caps how much of a dirty repo's diff --verbose shows, so one large change cannot bury the summary.
const maxDiffLines = 100

// localChangeLines returns what --verbose shows for a repo skipped for local changes: one line listing untracked files,
// then git diff HEAD capped at maxDiffLines, ending with the command that shows the rest when it is cut off.
func localChangeLines(g pullGitter, absPath string) []string {
	repo := scm.Repo{Name: filepath.Base(absPath), HostPath: absPath}

	status, err := g.WorkingTreeStatus(repo)
	if err != nil {
		return []string{"could not read status: " + err.Error()}
	}
	var lines, untracked []string
	for _, line := range strings.Split(status, "\n") {
		if path, ok := strings.CutPrefix(line, "?? "); ok {
			untracked = append(untracked, path)
		}
	}
	if len(untracked) > 0 {
		lines = append(lines, fmt.Sprintf("untracked (%d): %s", len(untracked), strings.Join(untracked, ", ")))
	}

	diff, err := g.DiffHead(repo)
	if err != nil {
		return append(lines, "could not read diff: "+err.Error())
	}
	if diff == "" {
		return lines
	}
	diffLines := strings.SplitN(diff, "\n", maxDiffLines+1)
	if len(diffLines) > maxDiffLines {
		more := strings.Count(diffLines[maxDiffLines], "\n") + 1
		diffLines = diffLines[:maxDiffLines]
		for i, line := range diffLines {
			diffLines[i] = strings.Clone(line) // do not keep the full diff alive through the kept lines
		}
		diffLines = append(diffLines, fmt.Sprintf(`... diff truncated (%d more lines), run: git -C "%s" diff HEAD`, more, absPath))
	}
	return append(lines, diffLines...)
}
